package main

import (
	"fmt"
	"github.com/calvarado2004/go-testing-webapp/pkg/data"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

var pathToTemplates = "./templates/"

// Home is the handler for the home page
func (app *application) Home(w http.ResponseWriter, r *http.Request) {

	var td = make(map[string]any)

	if app.Session.Exists(r.Context(), "test") {
		message := app.Session.GetString(r.Context(), "test")
		td["test"] = message
		log.Printf("session exists, message: %v", message)
	} else {
		app.Session.Put(r.Context(), "test", "Hit this page at "+time.Now().UTC().String())
		log.Printf("session created, it was empty")
	}

	err := app.render(w, r, "home.page.gohtml", &TemplateData{
		Data: td,
	})
	if err != nil {
		log.Printf("error rendering template: %v", err)
	}

}

// Profile is the handler for the profile page
func (app *application) Profile(w http.ResponseWriter, r *http.Request) {

	var td = make(map[string]any)

	if app.Session.Exists(r.Context(), "test") {
		message := app.Session.GetString(r.Context(), "test")
		td["test"] = message
		log.Printf("session exists, message: %v", message)
	} else {
		app.Session.Put(r.Context(), "test", "Hit this page at "+time.Now().UTC().String())
		log.Printf("session created, it was empty")
	}

	err := app.render(w, r, "profile.page.gohtml", &TemplateData{})
	if err != nil {
		log.Printf("error rendering template: %v", err)
	}

}

type TemplateData struct {
	IP    string
	Data  map[string]any
	Error string
	Flash string
	User  data.User
}

// render is a helper function that parses a template file and writes the
func (app *application) render(w http.ResponseWriter, r *http.Request, t string, td *TemplateData) error {

	// parse the template from disk
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, t), path.Join(pathToTemplates, "base.layout.gohtml"))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return err
	}

	td.IP = app.ipFromContext(r.Context())

	td.Error = app.Session.PopString(r.Context(), "error")

	td.Flash = app.Session.PopString(r.Context(), "flash")

	if app.Session.Exists(r.Context(), "user") {
		td.User = app.Session.Get(r.Context(), "user").(data.User)
	}

	// write the template to the http.ResponseWriter
	err = parsedTemplate.Execute(w, td)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return err
	}

	return nil
}

// Login is the handler for the login page
func (app *application) Login(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		log.Printf("error parsing form: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// validate the form data
	form := NewForm(r.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		//redirect to login page
		app.Session.Put(r.Context(), "error", "invalid login credentials")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.DB.GetUserByEmail(email)
	if err != nil {
		app.Session.Put(r.Context(), "error", "invalid login credentials")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// authenticate the user
	if !app.authenticate(r, user, password) {
		app.Session.Put(r.Context(), "error", "invalid login credentials")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// prevent fixation attack
	_ = app.Session.RenewToken(r.Context())

	// store success message in session

	// redirect to profile in page
	app.Session.Put(r.Context(), "flash", "You've been logged in successfully!")
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)

}

// authenticate checks the provided password against the hashed password stored in the database for a specific user.
func (app *application) authenticate(r *http.Request, user *data.User, password string) bool {

	// Check whether the provided password matches the hashed password in the database.
	if valid, err := user.PasswordMatches(password); err != nil || !valid {
		return false
	}

	app.Session.Put(r.Context(), "user", user)

	return true
}

// UploadProfilePic is the handler for the upload profile pic page
func (app *application) UploadProfilePic(w http.ResponseWriter, r *http.Request) {

	// call a function that extracts a file from an upload
	files, err := app.UploadFiles(r, "./static/img")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.Session.Put(r.Context(), "error", "error uploading file")
		return
	}

	// get the user from the session
	user := app.Session.Get(r.Context(), "user").(data.User)

	// create a var of file data.UserImage
	var imageVar = data.UserImage{
		UserID:   user.ID,
		FileName: files[0].OriginalFileName,
	}

	// insert the user image into user_images
	_, err = app.DB.InsertUserImage(imageVar)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.Session.Put(r.Context(), "error", "error uploading file")
		return
	}

	// update the user's profile pic session variable "user"
	updatedUser, err := app.DB.GetUser(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.Session.Put(r.Context(), "error", "error uploading file")
		return
	}

	app.Session.Put(r.Context(), "user", updatedUser)

	// redirect to profile page
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)

}

// UploadedFile is a struct that holds the original file name and the file size
type UploadedFile struct {
	OriginalFileName string
	FileSize         int64
}

// UploadFiles is the handler for the upload files page
func (app *application) UploadFiles(r *http.Request, uploadDir string) ([]*UploadedFile, error) {

	var uploadedFiles []*UploadedFile

	err := r.ParseMultipartForm(int64(1024 * 1024 * 100))
	if err != nil {
		return nil, fmt.Errorf("error parsing multipart form, bigger than 100Mib: %v", err)
	}

	for _, fileHeaders := range r.MultipartForm.File {
		for _, hdr := range fileHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {

				var uploadedFile UploadedFile

				infile, err := hdr.Open()
				if err != nil {
					return nil, fmt.Errorf("error opening file: %v", err)
				}
				defer infile.Close()

				uploadedFile.OriginalFileName = hdr.Filename

				var outfile *os.File

				defer outfile.Close()

				if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.OriginalFileName)); err != nil {
					return nil, fmt.Errorf("error creating file: %v", err)
				} else {
					fileSize, err := io.Copy(outfile, infile)
					if err != nil {
						return nil, fmt.Errorf("error copying file: %v", err)
					}

					uploadedFile.FileSize = fileSize
				}

				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil

			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, err
			}
		}
	}

	return uploadedFiles, nil
}
