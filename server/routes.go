package server

import (
	"github.com/gorilla/mux"
	"github.com/han-so1omon/concierge.map/util"
)

func AddApproutes(route *mux.Router) {
	util.Logger.Info("Loading routes...")

	route.HandleFunc("/signin", SignInUser).Methods("POST")

	route.HandleFunc("/signup", SignUpUser).Methods("POST")

	route.HandleFunc("/signout", SignOutUser).Methods("POST")

	route.HandleFunc("/userInfo", GetUserInfo).Methods("GET")

	util.Logger.Info("Routes are loaded.")
}
