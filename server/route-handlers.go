package server

import (
	"context"
	"encoding/json"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"net/url"
	"time"
)

func SignInUser(response http.ResponseWriter, request *http.Request) {
	var loginRequest LoginParams
	var userRequested UserDetails
	var errorResponse = ErrorResponse{
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

		collection := client.Database("concierge").Collection("users")

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

		if !CheckPasswordHash(loginRequest.Password, userRequested.Password) {
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
		session, sessionErr := store.Get(request, "session")
		if sessionErr != nil {
			returnErrorResponse(response, request, errorResponse)
			return
		}

		user := &UserSessionInfo{
			Email:         userRequested.Email,
			Authenticated: userRequested.Authenticated,
		}
		session.Values["user"] = user

		err = session.Save(request, response)
		if err != nil {
			Logger.Info(err)
			returnErrorResponse(response, request, errorResponse)
			return
		}

		var successResponse = SuccessResponse{
			Code:    http.StatusOK,
			Message: "You are logged in",
			Response: SuccessfulLoginResponse{
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
	var existingUser UserDetails
	var registrationRequest RegistrationParams
	var errorResponse = ErrorResponse{
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
		var registrationResponse = SuccessfulSignupResponse{
			Email: registrationRequest.Email,
		}

		collection := client.Database("concierge").Collection("users")
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

		hashedPassword, _ := HashPassword(registrationRequest.Password)
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

		var successResponse = SuccessResponse{
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
	var signoutRequest SignoutParams
	var errorResponse = ErrorResponse{
		Code: http.StatusInternalServerError, Message: "It's not you it's me.",
	}

	decoder := json.NewDecoder(request.Body)
	decoderErr := decoder.Decode(&signoutRequest)
	defer request.Body.Close()

	if decoderErr != nil {
		returnErrorResponse(response, request, errorResponse)
		return
	}

	session, err := store.Get(request, "session")
	if err != nil {
		returnErrorResponse(response, request, errorResponse)
		return
	}

	session.Values["user"] = UserSessionInfo{}
	session.Options.MaxAge = -1

	err = session.Save(request, response)
	if err != nil {
		returnErrorResponse(response, request, errorResponse)
		return
	}

	var signoutResponse = SuccessfulSignoutResponse{
		Email: signoutRequest.Email,
	}

	var successResponse = SuccessResponse{
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
	var requestedUser UserDetails
	var errorResponse = ErrorResponse{
		Code: http.StatusInternalServerError, Message: "It's not you it's me.",
	}
	session, err := store.Get(request, "session")
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

	Logger.Infof("Requested user email: %s", user.Email)

	collection := client.Database("concierge").Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = collection.FindOne(ctx, bson.M{
		"email": user.Email,
	}).Decode(&requestedUser)

	defer cancel()

	if err != nil {
		returnErrorResponse(response, request, errorResponse)
		return
	}

	var userResponse = SuccessfulUserResponse{
		Email:    requestedUser.Email,
		Username: requestedUser.Username,
	}
	var successResponse = SuccessResponse{
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

func returnErrorResponse(response http.ResponseWriter, request *http.Request, errorMessage ErrorResponse) {
	httpResponse := &ErrorResponse{Code: errorMessage.Code, Message: errorMessage.Message}
	jsonResponse, err := json.Marshal(httpResponse)
	if err != nil {
		panic(err)
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(errorMessage.Code)
	response.Write(jsonResponse)
}

func getUser(s *sessions.Session) UserSessionInfo {
	val := s.Values["user"]
	var user = UserSessionInfo{}
	user, ok := val.(UserSessionInfo)
	if !ok {
		return UserSessionInfo{Authenticated: false}
	}
	return user
}
