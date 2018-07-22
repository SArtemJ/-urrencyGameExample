package libsteam

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

type MgoGameServer struct {
	Address   string
	APIPrefix string
	Router    *mux.Router
	Storage   *MongoStorage
}

type MgoGameServerConfig struct {
	address   string
	apiPrefix string
	Storage   *MongoStorage
}

type ReturnCurrency struct {
	Value float64 `json:"value"`
}

const (
	URLGetGames    = "http://api.steampowered.com/ISteamApps/GetAppList/v2"
	URLGetCostGame = "https://store.steampowered.com/api/appdetails/"

	//https://store.steampowered.com/api/appdetails/?appids=237110&cc=us&filters=price_overview&type=game
	//http://store.steampowered.com/api/appdetails?appids=57690&cc=us
)

func NewServer(cfg MgoGameServerConfig) *MgoGameServer {
	if cfg.address == "" {
		cfg.address = "localhost:8099"
	}
	if cfg.apiPrefix == "" {
		cfg.apiPrefix = "/api/"
	}
	server := &MgoGameServer{
		Address:   cfg.address,
		APIPrefix: cfg.apiPrefix,
		Router:    mux.NewRouter(),
		Storage:   cfg.Storage,
	}

	server.SetupRouter()
	return server
}

func (server *MgoGameServer) GetRouter() *mux.Router {
	return server.Router
}

func (server *MgoGameServer) SetupRouter() {
	server.Router = server.Router.PathPrefix(server.APIPrefix).Subrouter()
	Logger.Debugf(`API endpoint "%s"`, server.APIPrefix)

	server.Router.HandleFunc("/game", server.GetGameCost).Methods("POST")
	//server.Router.HandleFunc("/updall/", server.UpdateAllGames).Methods("GET")
	//server.Router.HandleFunc("/getdef/", server.GetDefaultGameCostFromSteam).Methods("GET")
}

func (server *MgoGameServer) Run() {
	Logger.Debugf(`MgoGameServer started on "%s"`, server.Address)
	server.Storage.Reset()
	if server.GetAllGamesSteam() == false {
		Logger.Debugw("Error init data about games - try again or check request address")
	}
	Logger.Debugw("Init game data to Mongo - ok")
	// server.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
	// 	t, err := route.GetPathTemplate()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fmt.Println(t)
	// 	return nil
	// })
	http.ListenAndServe(server.Address, server.Router)
}

func (server *MgoGameServer) GetGameCost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	gameID := r.Form.Get("appid")
	currency := r.Form.Get("currency")
	Logger.Debugw("POST request get cost game", "game id", gameID, "currency", currency)

	if server.GetDefaultGameCostFromSteam(gameID) == true {
		if app, ok := server.Storage.CheckAndReturnGameInDB(gameID); ok == true {
			basicCost := app.App.USD * 100.00 //cent
			costInBTC := server.GetDefaultCostApp_InBTC(basicCost)

			app.M.Lock()
			defer app.M.Unlock()
			switch currency {
			case "EUR":
				f := server.ConvertCost(basicCost, "BTCEUR") / 100.00
				app.App.EUR = FloatFixed(f)
				server.Storage.UpdateFiledByID(app.App.ID, "EUR", app.App.EUR)
			case "GBP":
				f := server.ConvertCost(basicCost, "BTCGBP") / 100.00
				app.App.GBP = FloatFixed(f)
				server.Storage.UpdateFiledByID(app.App.ID, "GBP", app.App.GBP)
			case "RUB":
				f := server.ConvertCost(basicCost, "BTCRUB") / 100.00
				app.App.RUB = FloatFixed(f)
				server.Storage.UpdateFiledByID(app.App.ID, "RUB", app.App.RUB)
			case "BTC":
				app.App.BTC = costInBTC
				server.Storage.UpdateFiledByID(app.App.ID, "BTC", app.App.BTC)
			case "USD":
				app.App.USD = basicCost
				server.Storage.UpdateFiledByID(app.App.ID, "USD", app.App.USD)
			}
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			json.NewEncoder(w).Encode(app.App)
			w.WriteHeader(http.StatusOK)
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
		io.WriteString(w, "No game info please try again")
		Logger.Debugw("Not exist game in Mongo DB")
	}
}

func (server *MgoGameServer) GetDefaultCostApp_InBTC(basicCostInUSD float64) float64 {
	costAppInBTC := 0.00
	if v, ok := server.RequestToCurrencyAPI("BTCUSD"); ok == true {
		costAppInBTC = basicCostInUSD / v //game cost in BTC
	}
	return costAppInBTC
}

func (server *MgoGameServer) ConvertCost(basicCostInUSD float64, typeCost string) float64 {
	result := 0.00
	if v, ok := server.RequestToCurrencyAPI("BTCUSD"); ok == true {
		sAppInBTC := basicCostInUSD / v //BTC in cent
		if newCost, ok := server.RequestToCurrencyAPI(typeCost); ok == true {
			//Cost in new type
			result = (sAppInBTC * newCost)
		}
	}
	return result
}

func (server *MgoGameServer) GetAllGamesSteam() bool {
	req, err := http.NewRequest(http.MethodGet, URLGetGames, nil)
	if err != nil {
		Logger.Debugw("Error create request to get data about Games")
		return false
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		Logger.Debugw("Error response to get data about games")
		return false
	}

	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Logger.Debugw("Error read esponse body with info about games")
		return false
	}

	var data SteamApps
	err = json.Unmarshal(b, &data)
	if err != nil {
		Logger.Debugw("Can't parse response body to struct Go - info about games")
		return false
	}

	for _, v := range data.Applist.Apps {
		v.ID = bson.NewObjectId()
		v.USD = 0
		v.EUR = 0
		v.GBP = 0
		v.RUB = 0
		v.BTC = 0
		if err := server.Storage.Db.C(server.Storage.Collection).Insert(v); err != nil {
			Logger.Debugw("Can't save data info about games in MongoDB", " appid - ", err)
			continue
		}
	}
	return true
}

func (server *MgoGameServer) GetDefaultGameCostFromSteam(AppID string) bool {
	done := false

	if game, ok := server.Storage.CheckAndReturnGameInDB(AppID); ok == true {
		game.M.Lock()
		defer game.M.Unlock()
		request := fmt.Sprintf(URLGetCostGame+"?appids=%s&cc=us&filters=price_overview&type=game", AppID)

		if b, ok := server.DoRequest("GET", request); ok == true {
			data := make(map[int]SteamAppPrice)
			err := json.Unmarshal(b, &data)
			if err != nil {
				Logger.Debugw("Can't parse response body to struct Go - game cost", err)
				return done
			}
			appIDInt, err := strconv.Atoi(AppID)
			if err != nil {
				Logger.Debugw("Bad id to request try again", err)
				return done
			}
			if _, ok := data[appIDInt]; ok {
				game.App.USD = float64(data[appIDInt].Data.Price.Final)
			} else {
				Logger.Debugw("Not exist id in map from JSON game cost", err)
				return done
			}
		}

		if ok := server.Storage.UpdateFiledByID(game.App.ID, "USD", game.App.USD); ok == true {
			Logger.Debugw("Defaulf cost by USD updated to game - " + game.App.Name)
			done = true
			return done
		} else {
			Logger.Debugw("Can't update cost by USD to game" + game.App.Name)
			return done
		}
	}
	return done
}

func (server *MgoGameServer) DoRequest(method, url string) ([]byte, bool) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		Logger.Debugw("Error create request with method", " - ", method)
		Logger.Debugw("Error create request with url", " - ", url)
		return nil, false
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		Logger.Debugw("Error response method", " - ", method)
		Logger.Debugw("Error response request", " - ", url)
		return nil, false
	}

	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Logger.Debugw("Error read esponse method", " - ", method)
		Logger.Debugw("Error read esponse url", " - ", url)
		return nil, false
	}
	return b, true
}

func (server *MgoGameServer) RequestToCurrencyAPI(typeCurrency string) (float64, bool) {
	url := fmt.Sprintf("http://localhost:8888/api/currency/%s", typeCurrency)

	var data ReturnCurrency
	var ok = false

	if b, ok := server.DoRequest("GET", url); ok == true {
		err := json.Unmarshal(b, &data)
		if err != nil {
			Logger.Debugw("Can't parse response body to struct Go - currency game")
			return 0.00, ok
		}
	}

	//cent's
	oneBTC := data.Value * 100.00
	ok = true
	return oneBTC, ok
}

func FloatFixed(num float64) float64 {
	output := math.Pow(10, float64(2))
	return float64(math.Round(num*output)) / output
}
