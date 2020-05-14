package auth

import (
	"encoding/gob"
	"encoding/json"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"reflect"

	"github.com/han-so1omon/concierge.map/data"
	"github.com/han-so1omon/concierge.map/util"
)

var Store *sessions.CookieStore

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type Keys struct {
	AuthKey       []byte
	EncryptionKey []byte
}

func InitSession(sessionKeys *Keys) {
	if reflect.DeepEqual(sessionKeys, (Keys{})) {
		util.Logger.Fatal("Session keys are unset")
	}

	Store = sessions.NewCookieStore(
		sessionKeys.AuthKey,
		sessionKeys.EncryptionKey,
	)

	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   24 * 60 * 60,
		HttpOnly: true,
	}

	gob.Register(data.UserSessionInfo{})
}

func CheckUser(next http.Handler) http.Handler {
	var errorResponse = data.ErrorResponse{
		Code: http.StatusInternalServerError, Message: "It's not you it's me.",
	}
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		session, err := Store.Get(request, "session")
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
		next.ServeHTTP(response, request)
	})
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
