package main

import (
	"encoding/gob"
	"github.com/gorilla/handlers"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"net/http"
	"os"

	"github.com/han-so1omon/concierge.map/auth"
	"github.com/han-so1omon/concierge.map/data/db"
	"github.com/han-so1omon/concierge.map/server"
	"github.com/han-so1omon/concierge.map/util"
)

const defaultPort = "8000"

func main() {
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

	auth.InitSession(&sessionKeys)

	db.ConnectDatabase()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	router := httprouter.New()
	server.AddApproutes(router)

	headersOk := handlers.AllowedHeaders([]string{
		"Accept",
		"Accept-Encoding",
		"Accept-Language",
		"Cache-Control",
		"Connection",
		"Content-Length",
		"Content-Type",
		"Host",
		"Origin",
		"Pragma",
		"Referer",
		"Set-Cookie",
		"User-Agent",
		"X-Requested-With",
	})
	originsOk := handlers.AllowedOrigins([]string{
		"http://192.168.1.23:8080",
		"http://192.168.1.37:8080",
		"http://errcsool.com",
		"https://errcsool.com",
	})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	credentialsOk := handlers.AllowCredentials()

	util.Logger.Infof("Connect to http://0.0.0.0:%s/ for GraphQL playground", port)
	util.Logger.Fatal(http.ListenAndServe(
		"0.0.0.0:"+port,
		handlers.CORS(originsOk, headersOk, methodsOk, credentialsOk)(router)),
	)
}
