package libcurrency

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/nicovogelaar/go-bitcoinaverage/bitcoinaverage"
)

type CurrencyServer struct {
	Address   string
	APIPrefix string
	PublicKey string
	SecretKey string

	Timer   *time.Timer
	Router  *mux.Router
	RClient *redis.Client
}

type CurrencyServerConfig struct {
	address   string
	apiPrefix string
	timer     int64
	publicKey string
	secretKey string
}

func NewServer(cfg CurrencyServerConfig) *CurrencyServer {
	if cfg.address == "" {
		cfg.address = "/"
	}
	if cfg.apiPrefix == "" {
		cfg.apiPrefix = "/api/"
	}
	if cfg.timer == 0 {
		cfg.timer = 5
	}
	if cfg.publicKey == "" {
		cfg.publicKey = ""
	}
	if cfg.secretKey == "" {
		cfg.secretKey = ""
	}

	server := &CurrencyServer{
		Address:   cfg.address,
		APIPrefix: cfg.apiPrefix,
		PublicKey: cfg.publicKey,
		SecretKey: cfg.secretKey,
		Timer:     time.NewTimer(time.Duration(cfg.timer)),
		Router:    mux.NewRouter(),
	}

	server.SetupRouter()
	return server
}

func (server *CurrencyServer) SetupRouter() {
	server.Router = server.Router.PathPrefix(server.APIPrefix).Subrouter()
	//Logger

	server.Router.HandleFunc("/update/{type}", server.UpdateCurrency).Methods("PATCH")
	server.Router.HandleFunc("/currency/{type}", server.GetCurrency).Methods("GET")
	server.Router.HandleFunc("/updateall", server.UpdateAllCurrency).Methods("GET")
}

func (server *CurrencyServer) Run() {
	server.RedisConnection()
	Logger.Infof(`Stream server started on "%s"`, server.Address)

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

func (server *CurrencyServer) UpdateCurrency(w http.ResponseWriter, r *http.Request) {

}

func (server *CurrencyServer) GetCurrency(w http.ResponseWriter, r *http.Request) {

}

func (server *CurrencyServer) UpdateAllCurrency(w http.ResponseWriter, r *http.Request) {
	bitcoinClient := bitcoinaverage.NewClient(server.PublicKey, server.SecretKey)
	priceDataService := bitcoinaverage.NewPriceDataService(bitcoinClient)

	priceData, err := priceDataService.GetTickerDataBySymbol(bitcoinaverage.SymbolSetGlobal, "BTCUSD")
	if err != nil {
		Logger.Debugw("No currency data to save or bad request to bitcoinaverage")
	} else {
		server.SetRValue("BTCUSD", strconv.FormatFloat(priceData.Ask, 'f', -1, 64))
	}
	//fmt.Println(server.GetRValue("BTCUSD"))
}

func (server *CurrencyServer) GetRouter() *mux.Router {
	return server.Router
}

func (server *CurrencyServer) RedisConnection() {
	server.RClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := server.RClient.Ping().Result()
	if err != nil {
		Logger.Debugw("No connection to Redis")
		return
	}
	Logger.Debugw("Redis connection - ok")
}

func (server *CurrencyServer) SetRValue(key string, value string) {
	err := server.RClient.Set(key, value, 0).Err()
	if err != nil {
		panic(err)
	}
}

func (server *CurrencyServer) GetRValue(key string) string {
	val, err := server.RClient.Get(key).Result()
	if err != nil {
		//panic(err)
	}
	return val
}
