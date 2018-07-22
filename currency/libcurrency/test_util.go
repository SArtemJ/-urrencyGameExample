package libcurrency

var testApp *Application

func GetTestApp(cfg map[string]interface{}) *Application {
	if testApp == nil {
		testApp = NewApplication()
		testApp.Configure("currency_test")
		testApp.Init()
	}
	return testApp
}

func GetTestServer() *CurrencyServer {
	return GetTestApp(nil).Server
}
