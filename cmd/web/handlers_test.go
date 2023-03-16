package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"github.com/calvarado2004/go-testing-webapp/pkg/data"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
)

// Test_application_handlers tests the handlers
func Test_application_handlers(t *testing.T) {

	// create a slice of anonymous structs containing the name of the test, the URL path to
	var theTests = []struct {
		name                    string
		url                     string
		expectedStatusCode      int
		expectedURL             string
		expectedFirstStatusCode int
	}{
		{"home", "/", http.StatusOK, "/", http.StatusOK},
		{"404", "/fish", http.StatusNotFound, "/fish", http.StatusNotFound},
		{"profile", "/user/profile", http.StatusOK, "/", http.StatusTemporaryRedirect},
	}

	routes := app.routes()

	// create a test server using the routes
	ts := httptest.NewTLSServer(routes)

	// defer the closing of the test server until the test function has completed
	defer ts.Close()

	// create a new URL from the test server URL
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// loop through the slice of anonymous structs
	for _, tt := range theTests {
		resp, err := ts.Client().Get(ts.URL + tt.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != tt.expectedStatusCode {
			t.Errorf("%s: expected %d; got %d", tt.name, tt.expectedStatusCode, resp.StatusCode)
		}

		resp2, _ := client.Get(ts.URL + tt.url)
		if resp2.StatusCode != tt.expectedFirstStatusCode {
			t.Errorf("%s: expected %d; got %d", tt.name, tt.expectedFirstStatusCode, resp2.StatusCode)
		}

		if resp.Request.URL.Path != tt.expectedURL {
			t.Errorf("%s: expected final url %s; got %s", tt.name, tt.expectedURL, resp.Request.URL.Path)
		}

	}

}

// TestAppHomeOld tests the home page handler
func TestAppHomeOld(t *testing.T) {

	// create a new request for the home page
	req, _ := http.NewRequest("GET", "/", nil)

	// add a context and session to the request
	req = addContextAndSessionToRequest(req, app)

	// create a new response recorder
	rw := httptest.NewRecorder()

	handler := http.HandlerFunc(app.Home)

	handler.ServeHTTP(rw, req)

	// check the status code is what we expect
	if rw.Code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, rw.Code)
	}

	// read the response body
	body, _ := io.ReadAll(rw.Body)

	// check the response body is what we expect
	if !strings.Contains(string(body), "Your request came from") {
		t.Errorf("want %q; got %q", "Welcome to the home page", string(body))
	}

}

// TestAppHome tests the home page handler using a table driven test
func TestAppHome(t *testing.T) {

	var theTests = []struct {
		name            string
		putInSession    string
		expectedContent string
	}{
		{"first visit", "", "Your request came from"},
		{"second visit", "hello, world!", "hello, world!"},
	}

	for _, tt := range theTests {

		// create a new request for the home page
		req, _ := http.NewRequest("GET", "/", nil)

		// add a context and session to the request
		req = addContextAndSessionToRequest(req, app)

		_ = app.Session.Destroy(req.Context())

		if tt.putInSession != "" {

			// add a value to the session if empty
			app.Session.Put(req.Context(), "test", tt.putInSession)
		}

		// create a new response recorder
		rw := httptest.NewRecorder()

		handler := http.HandlerFunc(app.Home)

		handler.ServeHTTP(rw, req)

		// check the status code is what we expect
		if rw.Code != http.StatusOK {
			t.Errorf("want %d; got %d", http.StatusOK, rw.Code)
		}

		// read the response body
		body, _ := io.ReadAll(rw.Body)

		// check the response body is what we expect
		if !strings.Contains(string(body), tt.expectedContent) {
			t.Errorf("want %q; got %q", tt.expectedContent, string(body))
		}

	}
}

// TestApp_renderWithBadTemplate tests the render function with a bad template
func TestApp_renderWithBadTemplate(t *testing.T) {

	// set template path to a bad path
	pathToTemplates = "./testdata"

	// create a new request for the home page
	req, _ := http.NewRequest("GET", "/", nil)

	// add a context and session to the request
	req = addContextAndSessionToRequest(req, app)

	// create a new response recorder
	rw := httptest.NewRecorder()

	error := app.render(rw, req, "bad.page.gothtml", &TemplateData{})

	if error == nil {
		t.Error("expected an error to be returned")
	}

	// set template path back to the correct path
	pathToTemplates = "./../../templates/"

}

// getCtx returns a context with a value added
func getCtx(req *http.Request) context.Context {

	ctx := context.WithValue(req.Context(), contextUserKey, "unknown")

	return ctx
}

// addContextAndSessionToRequest adds a context and session to the request
func addContextAndSessionToRequest(req *http.Request, app application) *http.Request {

	req = req.WithContext(getCtx(req))

	// add the session to the context
	ctx, _ := app.Session.Load(req.Context(), req.Header.Get("X-Session"))

	return req.WithContext(ctx)
}

// Test_app_Login tests the login handler
func Test_app_Login(t *testing.T) {

	// create a new request for the login page
	var theTests = []struct {
		name               string
		postedData         url.Values
		expectedStatusCode int
		expectedLocation   string
	}{
		{
			name: "valid credentials",
			postedData: url.Values{
				"email":    {"admin@example.com"},
				"password": {"secret"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/user/profile",
		},
		{
			name: "user not found",
			postedData: url.Values{
				"email":    {"me@here.com"},
				"password": {"secret"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/",
		},
		{
			name: "missing credentials",
			postedData: url.Values{
				"email":    {""},
				"password": {""},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/",
		},
		{
			name: "wrong credentials",
			postedData: url.Values{
				"email":    {"admin@example.com"},
				"password": {"wrong"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/",
		},
		{
			name: "missing password",
			postedData: url.Values{
				"email":    {"admin@example.com"},
				"password": {""},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/",
		},
		{
			name: "missing email",
			postedData: url.Values{
				"email":    {""},
				"password": {"secret"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/",
		},
	}

	// loop through the tests
	for _, tt := range theTests {
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(tt.postedData.Encode()))
		req = addContextAndSessionToRequest(req, app)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rw := httptest.NewRecorder()
		handler := http.HandlerFunc(app.Login)
		handler.ServeHTTP(rw, req)

		// check the status code is what we expect
		if rw.Code != tt.expectedStatusCode {
			t.Errorf("%s: expected %d; got %d", tt.name, tt.expectedStatusCode, rw.Code)
		}

		// check the location header is what we expect using header.Contains()
		if rw.Header().Get("Location") == "" {
			t.Errorf("%s expected a location header to be set", tt.name)
		}

		// check the location header is what we expect using header.Get()
		if rw.Header().Get("Location") != tt.expectedLocation {
			t.Errorf("%s: expected %s; got %s", tt.name, tt.expectedLocation, rw.Header().Get("Location"))
		}

	}

}

// Test_app_UploadFiles tests the upload files handler
func Test_app_UploadFiles(t *testing.T) {

	// set up pipes
	pr, pw := io.Pipe()

	// create a new writer of type *io.Writer
	writer := multipart.NewWriter(pw)

	// create a wait group and add 1 to it
	wg := sync.WaitGroup{}
	wg.Add(1)

	// simulate uploading a file using a goroutine and our writer
	go simulatePNGUpload("./testdata/img.png", writer, t, &wg)

	// read from the file which receives the data from the writer
	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	// call app.UploadFiles with the request and response recorder
	uploadedFiles, err := app.UploadFiles(request, "./testdata/uploads/")
	if err != nil {
		t.Error(err)
	}

	// perform the tests
	if _, err := os.Stat("./testdata/uploads/" + uploadedFiles[0].OriginalFileName); os.IsNotExist(err) {
		t.Error("file was not uploaded, file does not exist")
	}

	// clean up the files
	_ = os.Remove("./testdata/uploads/" + uploadedFiles[0].OriginalFileName)

}

// simulatePNGUpload simulates uploading a png file
func simulatePNGUpload(fileToUpload string, writer *multipart.Writer, t *testing.T, wg *sync.WaitGroup) {

	defer writer.Close()
	defer wg.Done()

	// create a new form file
	part, err := writer.CreateFormFile("file", fileToUpload)
	if err != nil {
		t.Error(err)
	}

	// open the file
	file, err := os.Open(fileToUpload)
	if err != nil {
		t.Error(err)
	}

	defer file.Close()

	// decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		t.Error("error decoding image")
	}

	// write the png to io.Writer
	err = png.Encode(part, img)
	if err != nil {
		t.Error("error encoding png")
	}

}

// Test_app_UploadProfilePic tests the upload profile pic handler
func Test_app_UploadProfilePic(t *testing.T) {

	uploadPath := "./testdata/uploads"
	filePath := "./testdata/img.png"

	// specify a field name for the form
	fieldName := "file"

	// create a bytes buffer to act as the request body

	body := new(bytes.Buffer)

	// create a multipart writer and write a multipart form to the buffer
	mw := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	if err != nil {
		t.Error(err)
	}

	defer file.Close()

	// create a form file
	fw, err := mw.CreateFormFile(fieldName, filePath)
	if err != nil {
		t.Error(err)
	}

	// copy the file data to the form file
	_, err = io.Copy(fw, file)
	if err != nil {
		t.Error(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/user/upload-profile-pic", body)
	req = addContextAndSessionToRequest(req, app)
	app.Session.Put(req.Context(), "user", data.User{ID: 1})
	req.Header.Add("Content-Type", mw.FormDataContentType())

	rw := httptest.NewRecorder()

	handler := http.HandlerFunc(app.UploadProfilePic)

	handler.ServeHTTP(rw, req)

	if rw.Result().StatusCode != http.StatusSeeOther {
		t.Errorf("expected %d; got %d", http.StatusSeeOther, rw.Result().StatusCode)
	}

	// clean up the files
	_ = os.Remove(uploadPath + "./testdata/uploads/img.png")

	mw.Close()
}
