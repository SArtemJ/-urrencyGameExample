package libsteam

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerStart(t *testing.T) {
	cfg := MgoGameServerConfig{
		address:   "localhost:5555",
		apiPrefix: "someapi",
	}
	server := NewServer(cfg)
	assert.Equal(t, "someapi", server.APIPrefix)
	assert.Equal(t, "localhost:5555", server.Address)
}

func TestDefaultLoadSteamApp(t *testing.T) {
	server := GetTestServer()

	server.GetAllGamesSteam()
	i, _ := server.Storage.Db.C(server.Storage.Collection).Count()
	assert.NotEqual(t, 0, i)
}

func TestGetInfoAbouGame(t *testing.T) {
	server := GetTestServer()

	server.GetAllGamesSteam()

	request := fmt.Sprintf("http://localhost:8099/api/aboutgame/%s", "20")
	req, _ := http.NewRequest("GET", request, nil)
	w := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var app AppsStruct
	_ = json.NewDecoder(w.Body).Decode(&app)
	log.Println(app)
	assert.Equal(t, 20, app.Appid)
	assert.Equal(t, "Team Fortress Classic", app.Name)
}

func TestGetGameCost(t *testing.T) {
	server := GetTestServer()

	server.GetAllGamesSteam()

	requestURL := fmt.Sprintf("http://localhost:8099/api/game")
	req, _ := http.NewRequest("POST", requestURL, nil)

	form := url.Values{}
	form.Add("appid", "20")
	form.Add("currency", "USD")
	req.PostForm = form
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var app AppsStruct
	_ = json.NewDecoder(w.Body).Decode(&app)
	log.Println(app)
	assert.Equal(t, 20, app.Appid)
	assert.Equal(t, "Team Fortress Classic", app.Name)
	assert.NotEqual(t, 0.00, app.USD)
}
