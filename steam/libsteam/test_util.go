package libsteam

var testApp *Application

func GetTestApp(cfg map[string]interface{}) *Application {
	if testApp == nil {
		testApp = NewApplication()
		testApp.Configure("steam_test")
		testApp.Init()
	}
	return testApp
}

func GetTestServer() *MgoGameServer {
	return GetTestApp(nil).Server
}
