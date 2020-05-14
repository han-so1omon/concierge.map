package server

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"

	"github.com/han-so1omon/concierge.map/auth"
	"github.com/han-so1omon/concierge.map/data/db"
	"github.com/han-so1omon/concierge.map/server/graph"
	"github.com/han-so1omon/concierge.map/server/graph/generated"
	"github.com/han-so1omon/concierge.map/util"
)

func AddApproutes(route *httprouter.Router) {
	util.Logger.Info("Loading routes...")

	// Middleware setup
	authMiddle := alice.New(auth.CheckUser)
	logMiddle := alice.New(util.LogRequest)

	signoutHandlerFunc := http.HandlerFunc(SignOutUser)
	// REST setup
	route.HandlerFunc(http.MethodPost, "/signin", SignInUser)
	route.HandlerFunc(http.MethodPost, "/signup", SignUpUser)
	route.Handler(http.MethodPost, "/signout", authMiddle.Then(signoutHandlerFunc))
	route.HandlerFunc(http.MethodGet, "/userInfo", GetUserInfo)

	// GraphQL setup
	c := generated.Config{Resolvers: &graph.Resolver{
		UserCollection:    db.Client.Database("concierge").Collection("users"),
		ProjectCollection: db.Client.Database("concierge").Collection("projects"),
	}}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(c))

	route.Handler(
		http.MethodGet,
		"/graphql",
		playground.Handler("GraphQL playground", "/query"),
		//authMiddle.Then(playground.Handler("GraphQL playground", "/query")),
		//logMiddle.Then(playground.Handler("GraphQL playground", "/query")),
	)
	route.Handler(
		http.MethodPost,
		"/query",
		//srv,
		//authMiddle.Then(srv),
		logMiddle.Then(srv),
	)

	util.Logger.Info("Routes are loaded.")
}
