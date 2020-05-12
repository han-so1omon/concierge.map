package auth

import (
	"encoding/gob"
	"github.com/gorilla/sessions"
	"github.com/han-so1omon/concierge.map/data"
	"github.com/han-so1omon/concierge.map/util"
	"golang.org/x/crypto/bcrypt"
	"reflect"
)

var store *sessions.CookieStore

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

	store = sessions.NewCookieStore(
		sessionKeys.AuthKey,
		sessionKeys.EncryptionKey,
	)

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   24 * 60 * 60,
		HttpOnly: true,
	}

	gob.Register(data.UserSessionInfo{})
}
