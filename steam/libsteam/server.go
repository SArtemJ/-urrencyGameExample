package libsteam

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	//server.Router.HandleFunc("/game/", server.GetCostOneGame).Methods("GET")
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
		game.M.RLock()

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
				game.App.USD = data[appIDInt].Data.Price.Final
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

		game.M.RUnlock()
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
