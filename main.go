package main

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/han-so1omon/concierge.map/auth"
	"github.com/han-so1omon/concierge.map/data/db"
	"github.com/han-so1omon/concierge.map/server"
	"github.com/han-so1omon/concierge.map/server/graph"
	"github.com/han-so1omon/concierge.map/server/graph/generated"
	"github.com/han-so1omon/concierge.map/util"

	"encoding/gob"
	//"github.com/gorilla/handlers"
	//"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"net/http"
	"os"
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

	var sessionKeys Keys
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

	c := generated.Config{Resolvers: &graph.Resolver{
		UserCollection:    db.Client.Database("concierge").Collection("users"),
		ProjectCollection: db.Client.Database("concierge").Collection("projects"),
	}}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(c))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	util.Logger.Infof("Connect to http://0.0.0.0:%s/ for GraphQL playground", port)
	util.Logger.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

/*
package main

import (
	"encoding/gob"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	util.Logger = logger.Sugar()
	err := godotenv.Load()
	if err != nil {
		util.Logger.Fatal("Error loading .env file")
	}

	var sessionKeys Keys
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

	util.Logger.Info("Server will start at http://0.0.0.0:8000/")

	ConnectDatabase()

	route := mux.NewRouter()

	AddApproutes(route)

	headersOk := handlers.AllowedHeaders([]string{
		"Accept",
		"Content-Type",
		"Origin",
		"X-Requested-With",
		"Set-Cookie",
	})
	originsOk := handlers.AllowedOrigins([]string{
		"http://192.168.1.23:8080",
		"http://192.168.1.37:8080",
		"http://errcsool.com",
		"https://errcsool.com",
	})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	credentialsOk := handlers.AllowCredentials()

	util.Logger.Fatal(http.ListenAndServe(
		"0.0.0.0:8000",
		handlers.CORS(originsOk, headersOk, methodsOk, credentialsOk)(route)),
	)
}
*/
