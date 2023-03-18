package main

import (
	"errors"
	"fmt"
	"github.com/calvarado2004/go-testing-webapp/pkg/data"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strings"
	"time"
)

const jwtTokenExpiry = time.Minute * 15
const refreshTokenExpiry = time.Hour * 24

// TokenPairs is a struct that holds the access and refresh tokens
type TokenPairs struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Claims is a custom claims type
type Claims struct {
	UserName string `json:"username"`
	jwt.RegisteredClaims
}

// getTokenFromHeaderAndVerify verifies the token and returns the claims
func (app *application) getTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {

	// Bearer token is in the format:
	// Bearer <token>

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
			return nil, errors.New(fmt.Sprintf("invalid signing method %v", token.Header["alg"]))
		}

		return []byte(app.JWTSecret), nil
	})
	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired") {
			app.errorJSON(w, errors.New("token is expired"))
			return "", nil, errors.New("token is expired")
		}

		app.errorJSON(w, err)
		return "", nil, err
	}

	// make sure the token was issued by us
	if claims.Issuer != app.Domain {
		app.errorJSON(w, errors.New("invalid token, wrong issuer"))
		return "", nil, errors.New("invalid token, wrong issuer")
	}

	// valid token
	return token, claims, nil

}

// generateTokenPair generates a new access and refresh token for the user
func (app *application) generateTokenPair(user *data.User) (TokenPairs, error) {

	// create a new token for the user
	token := jwt.New(jwt.SigningMethodHS256)

	// set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = user.FirstName + " " + user.LastName
	claims["sub"] = user.ID
	claims["aud"] = app.Domain
	claims["iss"] = app.Domain
	if user.IsAdmin == 1 {
		claims["admin"] = true
	} else {
		claims["admin"] = false
	}
	claims["exp"] = time.Now().Add(jwtTokenExpiry).Unix()
	claims["iat"] = time.Now().Unix()

	// generate the token
	signedAccessToken, err := token.SignedString([]byte(app.JWTSecret))
	if err != nil {
		return TokenPairs{}, err
	}

	// create a new refresh token
	refreshToken := jwt.New(jwt.SigningMethodHS256)

	// set the claims
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)

	refreshClaims["sub"] = user.ID
	refreshClaims["exp"] = time.Now().Add(refreshTokenExpiry).Unix()

	signedRefreshToken, err := refreshToken.SignedString([]byte(app.JWTSecret))
	if err != nil {
		return TokenPairs{}, err
	}

	var tokenPairs TokenPairs
	tokenPairs.AccessToken = signedAccessToken
	tokenPairs.RefreshToken = signedRefreshToken

	return tokenPairs, nil
}
