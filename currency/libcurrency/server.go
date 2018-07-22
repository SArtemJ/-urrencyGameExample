package libcurrency

import (
	"encoding/json"
	"io"
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

	Ticker  *time.Ticker
	Router  *mux.Router
	RClient *redis.Client

	Currency map[string]float64
}

type CurrencyServerConfig struct {
	address   string
	apiPrefix string
	ticker    int64
	publicKey string
	secretKey string
}

type ReturnCurrency struct {
	Value float64 `json:"value"`
}

func NewServer(cfg CurrencyServerConfig) *CurrencyServer {
	if cfg.address == "" {
		cfg.address = "/"
	}
	if cfg.apiPrefix == "" {
		cfg.apiPrefix = "/api/"
	}
	if cfg.ticker == 0 {
		cfg.ticker = 5
	}
	if cfg.publicKey == "" {
		cfg.publicKey = "ODkzOGI3NTk3ODk1NGVmMDgzMDRiMWZkYTJiZDQzOTg"
	}
	if cfg.secretKey == "" {
		cfg.secretKey = "NTNlNDc2M2Y2ODJhNDViYmFlMjM5NGJmNDk2MTAxZDQwZGUyZWYxZTFmOTA0MTRjYWJkMGRmNTdiNTAzN2I4MQ"
	}

	server := &CurrencyServer{
		Address:   cfg.address,
		APIPrefix: cfg.apiPrefix,
		PublicKey: cfg.publicKey,
		SecretKey: cfg.secretKey,
		Ticker:    time.NewTicker(time.Minute * time.Duration(cfg.ticker)),
		Router:    mux.NewRouter(),
		Currency:  map[string]float64{"BTCUSD": 0.00, "BTCEUR": 0.00, "BTCGBP": 0.00, "BTCRUB": 0.00},
	}

	server.SetupRouter()
	return server
}

func (server *CurrencyServer) GetRouter() *mux.Router {
	return server.Router
}

func (server *CurrencyServer) SetupRouter() {
	server.Router = server.Router.PathPrefix(server.APIPrefix).Subrouter()
	Logger.Debugf(`API endpoint "%s"`, server.APIPrefix)

	server.Router.HandleFunc("/update/{type}", server.UpdateOneCurrency).Methods("PATCH")
	server.Router.HandleFunc("/currency/{type}", server.GetOneCurrency).Methods("GET")
	server.Router.HandleFunc("/currencyall", server.GetAllCurrency).Methods("GET")
	server.Router.HandleFunc("/updateall", server.UpdateAllCurrency).Methods("PATCH")
}

func (server *CurrencyServer) Run() {
	server.RedisConnection()
	go func() {
		for t := range server.Ticker.C {
			server.DoUpdateImmediately()
			Logger.Debugf(`Last update all currency "%s"`, t)
		}
	}()
	Logger.Debugf(`Stream server started on "%s"`, server.Address)
	http.ListenAndServe(server.Address, server.Router)
}

func (server *CurrencyServer) UpdateOneCurrency(w http.ResponseWriter, r *http.Request) {
	typeC := mux.Vars(r)["type"]
	if _, ok := server.Currency[typeC]; ok {
		if done := server.CurrencyUpdate(typeC); done == true {
			w.WriteHeader(http.StatusOK)
			resStr := "Value currency was updated " + typeC +
				" = " + strconv.FormatFloat(server.GetRValue(typeC), 'f', -1, 64)
			io.WriteString(w, resStr)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Not exist type of currency or server bitcoinaverage - return error")
		Logger.Debugw("Not exist type of currency for update - or server bitcoinaverage - return error", "err ", typeC)
	}
}

func (server *CurrencyServer) GetOneCurrency(w http.ResponseWriter, r *http.Request) {
	var resultC ReturnCurrency
	typeC := mux.Vars(r)["type"]
	if _, ok := server.Currency[typeC]; ok {
		resultC.Value = server.GetRValue(typeC)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode(resultC)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Bad request incorrect type currency")
		Logger.Debugw("Not exist type of currency for get method", "err ", typeC)
		return
	}
}

func (server *CurrencyServer) UpdateAllCurrency(w http.ResponseWriter, r *http.Request) {
	server.DoUpdateImmediately()
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "All currency was updated")
	Logger.Debugw("All currency was updated")
}

func (server *CurrencyServer) GetAllCurrency(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	allCurrency := map[string]float64{}
	for i, _ := range server.Currency {
		allCurrency[i] = server.GetRValue(i)
	}
	json.NewEncoder(w).Encode(allCurrency)
	w.WriteHeader(http.StatusOK)
}

func (server *CurrencyServer) DoUpdateImmediately() {
	for i, _ := range server.Currency {
		server.CurrencyUpdate(i)
	}
}

func (server *CurrencyServer) CurrencyUpdate(v string) bool {
	btcClient := bitcoinaverage.NewClient(server.PublicKey, server.SecretKey)
	btcDataService := bitcoinaverage.NewPriceDataService(btcClient)
	btcData, err := btcDataService.GetTickerDataBySymbol(bitcoinaverage.SymbolSetGlobal, v)
	if err != nil {
		Logger.Debugw("No currency data to save or bad request to bitcoinaverage")
		return false
	} else {
		server.SetRValue(v, btcData.Ask)
		return true
	}
}

//redis
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

	server.RClient.FlushAll()
	for i, v := range server.Currency {
		server.SetRValue(i, v)
	}
	Logger.Debugw("Redis connection - ok")
}

func (server *CurrencyServer) SetRValue(key string, value float64) {
	err := server.RClient.Set(key, value, 0).Err()
	if err != nil {
		Logger.Debugw("Can't set value to Redis")
		return
	}
}

func (server *CurrencyServer) GetRValue(key string) float64 {
	val, err := server.RClient.Get(key).Float64()
	if err != nil {
		Logger.Debugw("Can't get value from Redis")
		return 0.00
	}
	return val
}
