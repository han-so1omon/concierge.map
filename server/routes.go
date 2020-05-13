package server

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/julienschmidt/httprouter"
	"net/http"
	//"github.com/gorilla/handlers"

	"github.com/han-so1omon/concierge.map/data/db"
	"github.com/han-so1omon/concierge.map/server/graph"
	"github.com/han-so1omon/concierge.map/server/graph/generated"
	"github.com/han-so1omon/concierge.map/util"
)

func AddApproutes(route *httprouter.Router) {
	util.Logger.Info("Loading routes...")

	// REST setup
	route.HandlerFunc(http.MethodPost, "/signin", SignInUser)

	route.HandlerFunc(http.MethodPost, "/signup", SignUpUser)

	route.HandlerFunc(http.MethodPost, "/signout", SignOutUser)

	route.HandlerFunc(http.MethodGet, "/userInfo", GetUserInfo)

	// GraphQL setup
	c := generated.Config{Resolvers: &graph.Resolver{
		UserCollection:    db.Client.Database("concierge").Collection("users"),
		ProjectCollection: db.Client.Database("concierge").Collection("projects"),
	}}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(c))

	route.Handler(http.MethodGet, "/", playground.Handler("GraphQL playground", "/query"))
	route.Handler(http.MethodPost, "/query", srv)

	util.Logger.Info("Routes are loaded.")
}
