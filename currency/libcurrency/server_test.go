package libcurrency

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerStart(t *testing.T) {
	server := GetTestServer()

	server.RedisConnection()
	str, err := server.RClient.Ping().Result()
	if err != nil {
		Logger.Debugw("No connection to Redis")
		return
	}
	assert.Equal(t, "PONG", str)
	assert.Equal(t, 4, len(server.Currency))
}

func TestUpdateAllCurrency(t *testing.T) {
	server := GetTestServer()
	server.RedisConnection()

	request := fmt.Sprintf("http://localhost:8888/api/updateall")
	req, _ := http.NewRequest("PATCH", request, nil)
	w := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	value := server.GetRValue("BTCUSD")
	assert.NotEqual(t, float64(0), value)

	value = server.GetRValue("BTCEUR")
	assert.NotEqual(t, float64(0), value)

	value = server.GetRValue("BTCGBP")
	assert.NotEqual(t, float64(0), value)

	value = server.GetRValue("BTCRUB")
	assert.NotEqual(t, float64(0), value)
}

func TestUpdateOneCurrency(t *testing.T) {
	server := GetTestServer()
	server.RedisConnection()

	request := fmt.Sprintf("http://localhost:8888/api/update/BTCRUB")
	req, _ := http.NewRequest("PATCH", request, nil)
	w := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	value := server.GetRValue("BTCRUB")
	assert.NotEqual(t, float64(0), value)

	server.SetRValue("BTCRUB", 155.55)
	value = server.GetRValue("BTCRUB")
	assert.Equal(t, 155.55, value)
}

func TestGetOneCurrency(t *testing.T) {
	server := GetTestServer()
	server.RedisConnection()

	request := fmt.Sprintf("http://localhost:8888/api/updateall")
	req, _ := http.NewRequest("PATCH", request, nil)
	w := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	request = fmt.Sprintf("http://localhost:8888/api/currency/BTCGBP")
	req, _ = http.NewRequest("GET", request, nil)
	w = httptest.NewRecorder()
	server.GetRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var rc ReturnCurrency
	_ = json.NewDecoder(w.Body).Decode(&rc)
	assert.NotEqual(t, float64(0), rc.Value)
}

func TestGetAllCurrency(t *testing.T) {
	server := GetTestServer()
	server.RedisConnection()

	request := fmt.Sprintf("http://localhost:8888/api/updateall")
	req, _ := http.NewRequest("PATCH", request, nil)
	w := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	request = fmt.Sprintf("http://localhost:8888/api/currencyall")
	req, _ = http.NewRequest("GET", request, nil)
	w = httptest.NewRecorder()
	server.GetRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var rc map[string]float64
	_ = json.NewDecoder(w.Body).Decode(&rc)
	assert.Equal(t, 4, len(rc))

	for _, v := range rc {
		assert.NotEqual(t, float64(0), v)
	}
}
