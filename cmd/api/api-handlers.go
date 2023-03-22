package main

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"time"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// authenticate handles the authentication process for existing users.
func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	// read a json payload
	err := app.readJSON(w, r, &creds)
	if err != nil {
		app.errorJSON(w, errors.New("unable to parse json"), http.StatusUnauthorized)
		return
	}

	// look up the user by email address
	user, err := app.DB.GetUserByEmail(creds.Email)
	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	// check the password matches
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))
	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	// create a new token for the user
	tokenPairs, err := app.generateTokenPair(user)
	if err != nil {
		app.errorJSON(w, errors.New("unable to generate token"), http.StatusInternalServerError)
		return
	}

	// write the token to the response
	err = app.writeJSON(w, http.StatusOK, tokenPairs, "token")
	if err != nil {
		app.errorJSON(w, errors.New("unable to write json"), http.StatusInternalServerError)
		return
	}

}

// refresh handles the refresh process for existing users.
func (app *application) refresh(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	refreshToken := r.Form.Get("refresh_token")

	// look up the user by refresh token using claims
	claims := &Claims{}

	_, err = jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (any, error) {
		return []byte(app.JWTSecret), nil
	})

	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if time.Unix(claims.ExpiresAt.Unix(), 0).Sub(time.Now()) > 30*time.Second {
		app.errorJSON(w, errors.New("token is not expired yet"), http.StatusTooEarly)
		return
	}

	// get the user id from the claims
	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.DB.GetUser(userID)
	if err != nil {
		app.errorJSON(w, errors.New("unknown user"), http.StatusBadRequest)
		return
	}

	// create a new token for the user
	tokenPairs, err := app.generateTokenPair(user)
	if err != nil {
		app.errorJSON(w, errors.New("unable to generate token"), http.StatusInternalServerError)
		return
	}

	// set a cookie for the refresh token
	http.SetCookie(w, &http.Cookie{
		Name:     "__Host-refresh_token",
		Path:     "/",
		Value:    tokenPairs.RefreshToken,
		Expires:  time.Now().Add(refreshTokenExpiry),
		MaxAge:   int(refreshTokenExpiry.Seconds()),
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Domain:   "localhost",
		HttpOnly: true,
	})

	// write the token to the response
	err = app.writeJSON(w, http.StatusOK, tokenPairs, "token")
	if err != nil {
		app.errorJSON(w, errors.New("unable to write json"), http.StatusInternalServerError)
		return
	}

}

// allUsers handles the GET /v1/users request.
func (app *application) allUsers(w http.ResponseWriter, r *http.Request) {

}

// getUser handles the GET /v1/users/{id} request.
func (app *application) getUser(w http.ResponseWriter, r *http.Request) {

}

// updateUser handles the PUT /v1/users/{id} request.
func (app *application) updateUser(w http.ResponseWriter, r *http.Request) {

}

// deleteUser handles the DELETE /v1/users/{id} request.
func (app *application) deleteUser(w http.ResponseWriter, r *http.Request) {

}

// insertUser handles the POST /v1/users request.
func (app *application) insertUser(w http.ResponseWriter, r *http.Request) {

}
