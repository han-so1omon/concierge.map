package server

import (
	"context"
	"encoding/json"
	"github.com/gorilla/sessions"
	"github.com/han-so1omon/concierge.map/auth"
	"github.com/han-so1omon/concierge.map/data"
	"github.com/han-so1omon/concierge.map/data/db"
	"github.com/han-so1omon/concierge.map/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"net/url"
	"time"
)

func SignInUser(response http.ResponseWriter, request *http.Request) {
	var loginRequest data.LoginParams
	var userRequested data.UserDetails
	var errorResponse = data.ErrorResponse{
		Code: http.StatusInternalServerError, Message: "It's not you it's me.",
	}

	decoder := json.NewDecoder(request.Body)
	decoderErr := decoder.Decode(&loginRequest)
	defer request.Body.Close()

	if decoderErr != nil {
		returnErrorResponse(response, request, errorResponse)
		return
	}

	errorResponse.Code = http.StatusBadRequest
	if loginRequest.Email == "" {
		errorResponse.Message = "Email can't be empty"
		returnErrorResponse(response, request, errorResponse)
		return
	} else if loginRequest.Password == "" {
		errorResponse.Message = "Password can't be empty"
		returnErrorResponse(response, request, errorResponse)
		return
	} else {

		collection := db.Client.Database("concierge").Collection("users")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var err = collection.FindOne(ctx, bson.M{
			"email": loginRequest.Email,
		}).Decode(&userRequested)

		defer cancel()

		if err == mongo.ErrNoDocuments {
			errorResponse.Message = "No user with this email"
			returnErrorResponse(response, request, errorResponse)
			return
		}

		if !auth.CheckPasswordHash(loginRequest.Password, userRequested.Password) {
			errorResponse.Message = "Incorrect password"
			returnErrorResponse(response, request, errorResponse)
			return
		}

		if err != nil {
			returnErrorResponse(response, request, errorResponse)
			return
		}
		// Create new user session
		//TODO should session name be unique (e.g. prepend with email name)
		session, sessionErr := auth.Store.Get(request, "session")
		if sessionErr != nil {
			returnErrorResponse(response, request, errorResponse)
			return
		}

		user := &data.UserSessionInfo{
			Email:         userRequested.Email,
			Authenticated: userRequested.Authenticated,
		}
		session.Values["user"] = user

		err = session.Save(request, response)
		if err != nil {
			util.Logger.Info(err)
			returnErrorResponse(response, request, errorResponse)
			return
		}

		var successResponse = data.SuccessResponse{
			Code:    http.StatusOK,
			Message: "You are logged in",
			Response: data.SuccessfulLoginResponse{
				Email:    loginRequest.Email,
				Username: userRequested.Username,
			},
		}

		successJSONResponse, jsonError := json.Marshal(successResponse)

		if jsonError != nil {
			returnErrorResponse(response, request, errorResponse)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		response.Write(successJSONResponse)
	}
}

func SignUpUser(response http.ResponseWriter, request *http.Request) {
	var existingUser data.UserDetails
	var registrationRequest data.RegistrationParams
	var errorResponse = data.ErrorResponse{
		Code: http.StatusInternalServerError, Message: "It's not you it's me.",
	}

	decoder := json.NewDecoder(request.Body)
	decoderErr := decoder.Decode(&registrationRequest)
	defer request.Body.Close()

	if decoderErr != nil {
		returnErrorResponse(response, request, errorResponse)
		return
	}
	errorResponse.Code = http.StatusBadRequest
	if registrationRequest.Username == "" {
		errorResponse.Message = "Username can't be empty"
		returnErrorResponse(response, request, errorResponse)
		return
	} else if registrationRequest.Email == "" {
		errorResponse.Message = "Email can't be empty"
		returnErrorResponse(response, request, errorResponse)
		return
	} else if registrationRequest.Password == "" {
		errorResponse.Message = "Password be empty"
		returnErrorResponse(response, request, errorResponse)
		return
	} else {
		var registrationResponse = data.SuccessfulSignupResponse{
			Email: registrationRequest.Email,
		}

		collection := db.Client.Database("concierge").Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Check if email already exists
		mongoErr := collection.FindOne(ctx, bson.M{
			"email": registrationRequest.Email,
		}).Decode(&existingUser)

		if mongoErr != nil && mongoErr != mongo.ErrNoDocuments {
			returnErrorResponse(response, request, errorResponse)
			return
		}
		if existingUser.Email == registrationRequest.Email {
			errorResponse.Message = "User with this email already exists"
			returnErrorResponse(response, request, errorResponse)
			return
		}

		hashedPassword, _ := auth.HashPassword(registrationRequest.Password)
		_, databaseErr := collection.InsertOne(ctx, bson.M{
			"email":         registrationRequest.Email,
			"password":      hashedPassword,
			"username":      registrationRequest.Username,
			"authenticated": false,
		})

		if databaseErr != nil {
			returnErrorResponse(response, request, errorResponse)
			return
		}

		var successResponse = data.SuccessResponse{
			Code:     http.StatusOK,
			Message:  "You are registered, login to continue",
			Response: registrationResponse,
		}

		successJSONResponse, jsonError := json.Marshal(successResponse)

		if jsonError != nil {
			returnErrorResponse(response, request, errorResponse)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(successResponse.Code)
		response.Write(successJSONResponse)
	}
}

func SignOutUser(response http.ResponseWriter, request *http.Request) {
	var signoutRequest data.SignoutParams
	var errorResponse = data.ErrorResponse{
		Code: http.StatusInternalServerError, Message: "It's not you it's me.",
	}

	decoder := json.NewDecoder(request.Body)
	decoderErr := decoder.Decode(&signoutRequest)
	defer request.Body.Close()

	if decoderErr != nil {
		returnErrorResponse(response, request, errorResponse)
		return
	}

	session, err := auth.Store.Get(request, "session")
	if err != nil {
		returnErrorResponse(response, request, errorResponse)
		return
	}

	session.Values["user"] = data.UserSessionInfo{}
	session.Options.MaxAge = -1

	err = session.Save(request, response)
	if err != nil {
		returnErrorResponse(response, request, errorResponse)
		return
	}

	var signoutResponse = data.SuccessfulSignoutResponse{
		Email: signoutRequest.Email,
	}

	var successResponse = data.SuccessResponse{
		Code:     http.StatusOK,
		Message:  "You have signed out",
		Response: signoutResponse,
	}

	successJSONResponse, jsonError := json.Marshal(successResponse)
	if jsonError != nil {
		returnErrorResponse(response, request, errorResponse)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(successResponse.Code)
	response.Write(successJSONResponse)
}

func GetUserInfo(response http.ResponseWriter, request *http.Request) {
	var requestedUser data.UserDetails
	var errorResponse = data.ErrorResponse{
		Code: http.StatusInternalServerError, Message: "It's not you it's me.",
	}
	session, err := auth.Store.Get(request, "session")
	if err != nil {
		returnErrorResponse(response, request, errorResponse)
		return
	}
	user := getUser(session)

	if user.Email == "" {
		errorResponse.Message = "No active user session"
		returnErrorResponse(response, request, errorResponse)
		return
	}

	util.Logger.Infof("Requested user email: %s", user.Email)

	collection := db.Client.Database("concierge").Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = collection.FindOne(ctx, bson.M{
		"email": user.Email,
	}).Decode(&requestedUser)

	defer cancel()

	if err != nil {
		returnErrorResponse(response, request, errorResponse)
		return
	}

	var userResponse = data.SuccessfulUserResponse{
		Email:    requestedUser.Email,
		Username: requestedUser.Username,
	}
	var successResponse = data.SuccessResponse{
		Code:     http.StatusOK,
		Message:  "Check it out!",
		Response: userResponse,
	}

	successJSONResponse, jsonError := json.Marshal(successResponse)

	if jsonError != nil {
		returnErrorResponse(response, request, errorResponse)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.Write(successJSONResponse)
}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func returnErrorResponse(response http.ResponseWriter, request *http.Request, errorMessage data.ErrorResponse) {
	httpResponse := &data.ErrorResponse{Code: errorMessage.Code, Message: errorMessage.Message}
	jsonResponse, err := json.Marshal(httpResponse)
	if err != nil {
		panic(err)
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(errorMessage.Code)
	response.Write(jsonResponse)
}

func getUser(s *sessions.Session) data.UserSessionInfo {
	val := s.Values["user"]
	var user = data.UserSessionInfo{}
	user, ok := val.(data.UserSessionInfo)
	if !ok {
		return data.UserSessionInfo{Authenticated: false}
	}
	return user
}
