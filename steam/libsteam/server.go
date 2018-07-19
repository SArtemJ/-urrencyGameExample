package libsteam

import (
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

	server.Router.HandleFunc("/game", server.CreateRandomGameDB).Methods("GET")
	// server.Router.HandleFunc("/currency/{type}", server.GetOneCurrency).Methods("GET")
	// server.Router.HandleFunc("/currencyall", server.GetAllCurrency).Methods("GET")
	// server.Router.HandleFunc("/updateall", server.UpdateAllCurrency).Methods("GET")
}

func (server *MgoGameServer) Run() {
	Logger.Debugf(`MgoGameServer started on "%s"`, server.Address)

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

func (server *MgoGameServer) CreateRandomGameDB(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var randGame GameInfo
	// if err := json.NewDecoder(r.Body).Decode(&randGame); err != nil {
	// 	Logger.Debugw("Nope")
	// 	return
	// }

	randGame.ID = bson.NewObjectId()
	randGame.Name = "Something"
	randGame.Cost = 100.00
	if err := server.Storage.Db.C(server.Storage.Collection).Insert(randGame); err != nil {
		Logger.Debugw("OOPS")
		return
	}
}
