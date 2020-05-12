package server

import (
	"bytes"
	"encoding/gob"
	"github.com/gorilla/mux"
	"github.com/han-so1omon/concierge.map/auth"
	"github.com/han-so1omon/concierge.map/util"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"testing"
)

var route *mux.Router

func TestMain(m *testing.M) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	util.Logger = logger.Sugar()
	err := godotenv.Load()
	if err != nil {
		util.Logger.Fatal("Error loading .env file")
	}

	var sessionKeys auth.Keys
	sessionKeysFile := os.Getenv("SESSION_KEYS_FILE")
	f, fileErr := os.OpenFile(sessionKeysFile, os.O_RDONLY, 0664)
	defer f.Close()
	if fileErr != nil {
		util.Logger.Fatal("Error opening session keys file")
	}
	dec := gob.NewDecoder(f)
	if fileErr = dec.Decode(&sessionKeys); fileErr != nil {
		util.Logger.Fatal("Error loading session keys")
	}

	initSession(&sessionKeys)

	ConnectDatabase()
	route = mux.NewRouter()

	AddApproutes(route)

	os.Exit(m.Run())
}

func TestCookies(t *testing.T) {
	t.Run("checking on cookies", func(t *testing.T) {
		dataJSON := []byte(`{"email": "x@x.com", "password": "x"}`)
		req, _ := http.NewRequest("POST", "/signin", bytes.NewBuffer(dataJSON))
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		req.Header.Set("ACCEPT", "application/json")
		recorder := httptest.NewRecorder()
		route.ServeHTTP(recorder, req)
		resp := recorder.Result()
		dump, _ := httputil.DumpResponse(resp, true)
		util.Logger.Infof("%q", dump)
	})
}
