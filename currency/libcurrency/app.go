package libcurrency

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	AppName = "currency"
)

var (
	Logger = zap.S()
)

type Application struct {
	cfg    *viper.Viper
	Server *CurrencyServer

	listenAddr        string
	serverAPIEndpoint string
	pubKey            string
	secretKey         string
	tickerValue       int

	rootCmd *cobra.Command
}

func NewApplication() *Application {
	app := Application{}
	return &app
}

func (app *Application) InitCommands() {

	app.rootCmd = &cobra.Command{
		Use:   "currency",
		Short: "currency API",
		Long:  "currency info API",
		Run: func(cmd *cobra.Command, args []string) {
			app.Init()
			app.Server.Run()
		},
	}

	app.rootCmd.PersistentFlags().StringVarP(&app.listenAddr, "service_address", "l", "localhost:8888", "service address")
	app.rootCmd.PersistentFlags().StringVarP(&app.pubKey, "pub_k_bitcoinaverage", "p", "", "public key to bitcoinaverage")
	app.rootCmd.PersistentFlags().StringVarP(&app.secretKey, "secret_k_bitcoinaverage", "s", "", "secret key to bitcoinaverage")
	app.rootCmd.PersistentFlags().StringVarP(&app.serverAPIEndpoint, "api", "a", "", "API URL endpoint")
	app.rootCmd.PersistentFlags().IntVarP(&app.tickerValue, "ticker_value", "t", 5, "time to wait")
}

func (app *Application) InitConfig(configName, envPrefix string) {
	cfg := viper.New()

	cfg.SetEnvPrefix(envPrefix)
	cfg.AutomaticEnv()
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg.SetDefault("server.addr", "localhost:8888")
	cfg.BindPFlag("server.addr", app.rootCmd.PersistentFlags().Lookup("service_address"))
	cfg.SetDefault("server.apiPrefix", "")
	cfg.BindPFlag("server.apiPrefix", app.rootCmd.PersistentFlags().Lookup("api"))
	cfg.SetDefault("ticker.value", 5)
	cfg.BindPFlag("ticker.value", app.rootCmd.PersistentFlags().Lookup("ticker_value"))
	cfg.SetDefault("pub.key", "")
	cfg.BindPFlag("pub.key", app.rootCmd.PersistentFlags().Lookup("pub_k_bitcoinaverage"))
	cfg.SetDefault("secret.key", "")
	cfg.BindPFlag("secret.key", app.rootCmd.PersistentFlags().Lookup("secret_k_bitcoinaverage"))

	cfg.SetConfigName(configName)
	cfg.AddConfigPath("/etc/")
	cfg.AddConfigPath("$HOME/")
	cfg.AddConfigPath("./")

	app.cfg = cfg
}

func (app *Application) GetConfig() *viper.Viper {
	return app.cfg
}

func (app *Application) ConfigureLog() {

	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := config.Build()
	if err != nil {
		panic("Failed to initialize logger")
	}
	Logger = logger.Sugar()
}

func (app *Application) Configure(params ...string) {
	configName := AppName
	envName := AppName
	switch {
	case len(params) == 1:
		configName = params[0]
		envName = params[0]
	case len(params) > 1:
		configName = params[0]
		envName = params[1]
	}
	app.InitCommands()
	app.InitConfig(configName, envName)
	app.ConfigureLog()
}

func (app *Application) Init() {

	app.listenAddr = app.cfg.GetString("server.addr")
	app.Server = NewServer(CurrencyServerConfig{
		address:   app.cfg.GetString("server.addr"),
		apiPrefix: app.cfg.GetString("server.apiPrefix"),
		ticker:    app.cfg.GetInt64("ticker.value"),
	})
	//app.Server.Timer = time.NewTimer(time.Second * time.Duration(app.timerValue))
}

func (app *Application) Run() {
	if err := app.rootCmd.Execute(); err != nil {
		panic(err)
	}
}
