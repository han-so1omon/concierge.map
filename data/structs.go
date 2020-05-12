package data

import (
	jwt "github.com/dgrijalva/jwt-go"
)

type ErrorResponse struct {
	Code    int
	Message string
}

type SuccessResponse struct {
	Code     int
	Message  string
	Response interface{}
}

type Claims struct {
	Email string
	jwt.StandardClaims
}

type RegistrationParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignoutParams struct {
	Email string `json:"email"`
}

type SuccessfulLoginResponse struct {
	Email    string
	Username string
}

type SuccessfulSignupResponse struct {
	Email string
}

type SuccessfulSignoutResponse struct {
	Email string
}

type SuccessfulUserResponse struct {
	Email    string
	Username string
}

type UserDetails struct {
	Email         string
	Password      string
	Username      string
	Authenticated bool
}

type UserSessionInfo struct {
	Email         string
	Authenticated bool
}
