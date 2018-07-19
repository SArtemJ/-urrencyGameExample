package libsteam

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

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
	URLGetCostGame = "http://store.steampowered.com/api/appdetails?appids="
	Language       = "l=en"
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

	server.Router.HandleFunc("/gameprice", server.GamePrice).Methods("POST")
	// server.Router.HandleFunc("/currency/{type}", server.GetOneCurrency).Methods("GET")
	// server.Router.HandleFunc("/currencyall", server.GetAllCurrency).Methods("GET")
	// server.Router.HandleFunc("/updateall", server.UpdateAllCurrency).Methods("GET")
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

type PriceOverview struct {
	Currency        string `json:"currency"`
	Initial         int    `json:"initial"`
	Final           int    `json:"final"`
	DiscountPercent int    `json:"discount_percent"`
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
		v.Cost = 0.00
		if err := server.Storage.Db.C(server.Storage.Collection).Insert(v); err != nil {
			Logger.Debugw("Can't save data info about games in MongoDB", " appid - ", err)
			continue
		}
	}
	return true
}

func (server *MgoGameServer) GamePrice(w http.ResponseWriter, r *http.Request) {

	//request := fmt.Sprintf(URLGetCostGame+"&%s"+"cc=%s&"+Language, "57690", "us")
	request := "https://store.steampowered.com/api/appdetails/?appids=237110&cc=us&filters=price_overview&type=game"
	req, err := http.NewRequest(http.MethodGet, request, nil)
	if err != nil {
		Logger.Debugw("Error create request to get data about Games")
		//return false
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		Logger.Debugw("Error response to get data about games")
		//return false
	}

	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Logger.Debugw("Error read esponse body with info about games")
		//return false
	}

	result := make(map[int]SteamAppPrice)
	json.Unmarshal(b, &result)

	json.NewEncoder(w).Encode(result)
	w.WriteHeader(http.StatusOK)
	//log.Printf("%v\n", result)

	//return true
}
