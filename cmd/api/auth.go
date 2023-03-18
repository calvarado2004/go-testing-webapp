package main

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strings"
	"time"
)

const jwtTokenExpiry = time.Minute * 15
const refreshTokenExpiry = time.Hour * 24

type TokenPairs struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	UserName string `json:"username"`
	jwt.RegisteredClaims
}

func (app *application) getTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {

	// add a header
	w.Header().Add("Vary", "Authorization")

	// get the authorization header
	authHeader := r.Header.Get("Authorization")

	// check if the header is empty
	if authHeader == "" {
		app.errorJSON(w, errors.New("no authorization header"))
		return "", nil, errors.New("no authorization header")
	}

	// split the header on spaces
	headersParts := strings.Split(authHeader, " ")
	if len(headersParts) != 2 {
		app.errorJSON(w, errors.New("invalid authorization header"))
		return "", nil, errors.New("invalid authorization header")
	}

	// check if the header is a bearer token
	if headersParts[0] != "Bearer" {
		app.errorJSON(w, errors.New("invalid authorization header"))
		return "", nil, errors.New("invalid authorization header")
	}

	token := headersParts[1]

	// verify the token
	claims := &Claims{}

	// parse the token
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		// validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		
		return []byte(app.JWTSecret), nil
	})
	if err != nil {
		app.errorJSON(w, err)
		return "", nil, err
	}

	return token, claims, nil

}
