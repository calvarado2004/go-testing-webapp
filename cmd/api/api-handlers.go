package main

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
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
