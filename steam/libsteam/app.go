package libsteam

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	AppName = "GamesApp"
)

var (
	Logger = zap.S()
)

type Application struct {
	cfg    *viper.Viper
	Server *MgoGameServer

	listenAddr        string
	serverAPIEndpoint string
	storageUri        string
	storageName       string

	rootCmd *cobra.Command
}

func NewApplication() *Application {
	app := Application{}
	return &app
}

func (app *Application) InitCommands() {

	app.rootCmd = &cobra.Command{
		Use:   "gameapp",
		Short: "game API",
		Long:  "game info API",
		Run: func(cmd *cobra.Command, args []string) {
			app.Init()
			app.Server.Run()
		},
	}

	app.rootCmd.PersistentFlags().StringVarP(&app.listenAddr, "service_address", "l", "localhost:8099", "service address")
	app.rootCmd.PersistentFlags().StringVarP(&app.storageUri, "storage_addr", "s", "localhost", "MongoDB server")
	app.rootCmd.PersistentFlags().StringVar(&app.storageName, "storage_name", "gamedb", "MongoDB database")
	app.rootCmd.PersistentFlags().StringVarP(&app.serverAPIEndpoint, "api", "a", "", "API URL endpoint")
}

func (app *Application) InitConfig(configName, envPrefix string) {
	cfg := viper.New()

	cfg.SetEnvPrefix(envPrefix)
	cfg.AutomaticEnv()
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg.SetDefault("server.addr", "localhost:8099")
	cfg.BindPFlag("server.addr", app.rootCmd.PersistentFlags().Lookup("service_address"))
	cfg.SetDefault("server.apiPrefix", "")
	cfg.BindPFlag("server.apiPrefix", app.rootCmd.PersistentFlags().Lookup("api"))
	cfg.SetDefault("storage.name", "gamedb")
	cfg.BindPFlag("storage.name", app.rootCmd.PersistentFlags().Lookup("storage_name"))
	cfg.SetDefault("storage.addr", "localhost")
	cfg.BindPFlag("storage.addr", app.rootCmd.PersistentFlags().Lookup("storage_addr"))

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
	storage := NewMongoStorage(app.cfg.GetString("storage.addr"), app.cfg.GetString("storage.name"))
	storage.Reset()

	app.Server = NewServer(MgoGameServerConfig{
		address:   app.cfg.GetString("server.addr"),
		apiPrefix: app.cfg.GetString("server.apiPrefix"),
		Storage:   storage,
	})
}

func (app *Application) Run() {
	if err := app.rootCmd.Execute(); err != nil {
		panic(err)
	}
}
