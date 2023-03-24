package main

import (
	"errors"
	"github.com/calvarado2004/go-testing-webapp/pkg/data"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"time"
)

type Credentials struct {
	Username string `json:"email"`
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
	user, err := app.DB.GetUserByEmail(creds.Username)
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

	w.WriteHeader(http.StatusOK)

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

	_, err = jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(app.JWTSecret), nil
	})

	if err != nil {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	if time.Unix(claims.ExpiresAt.Unix(), 0).Sub(time.Now()) > 30*time.Second {
		app.errorJSON(w, errors.New("token is not expired yet"), http.StatusTooEarly)
		return
	}

	// get the user id from the claims
	userID, _ := strconv.Atoi(claims.Subject)

	user, err := app.DB.GetUser(userID)
	if err != nil {
		app.errorJSON(w, errors.New("unknown user on db"), http.StatusNotFound)
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

	w.WriteHeader(http.StatusOK)
}

// allUsers handles the GET /v1/users request.
func (app *application) allUsers(w http.ResponseWriter, r *http.Request) {

	users, err := app.DB.AllUsers()
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	err = app.writeJSON(w, http.StatusOK, users, "users")
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

}

// getUser handles the GET /v1/users/{id} request.
func (app *application) getUser(w http.ResponseWriter, r *http.Request) {

	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.DB.GetUser(userID)
	if err != nil {
		app.errorJSON(w, err, http.StatusNotFound)
		return
	}

	err = app.writeJSON(w, http.StatusOK, user, "user")
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

}

// updateUser handles the PUT /v1/users/{id} request.
func (app *application) updateUser(w http.ResponseWriter, r *http.Request) {

	var user data.User

	err := app.readJSON(w, r, &user)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	err = app.DB.UpdateUser(user)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

// deleteUser handles the DELETE /v1/users/{id} request.
func (app *application) deleteUser(w http.ResponseWriter, r *http.Request) {

	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	err = app.DB.DeleteUser(userID)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// insertUser handles the POST /v1/users request.
func (app *application) insertUser(w http.ResponseWriter, r *http.Request) {

	var user data.User

	err := app.readJSON(w, r, &user)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	id, err := app.DB.InsertUser(user)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	user.ID = id

	err = app.writeJSON(w, http.StatusCreated, user, "user")
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

}
