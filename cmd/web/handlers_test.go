package main

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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
