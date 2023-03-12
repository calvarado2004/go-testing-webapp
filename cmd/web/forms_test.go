package main

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestForm_Has tests the Has() method of the Form type.
func TestForm_Has(t *testing.T) {

	// Test with no data
	form := NewForm(nil)

	has := form.Has("whatever")

	if has {
		t.Error("Has() returned true when it should have returned false")
	}

	// Test with valid data
	postedData := url.Values{}

	postedData.Add("a", "a")

	form = NewForm(postedData)

	has = form.Has("a")

	if !has {
		t.Error("Has() returned false when it should have returned true")
	}

}

// TestForm_Required tests the Required() method of the Form type.
func TestForm_Required(t *testing.T) {

	// Test with invalid data
	r := httptest.NewRequest("POST", "/whatever", nil)

	form := NewForm(r.PostForm)

	form.Required("a", "b", "c")

	if form.Valid() {
		t.Error("Form should not have been valid")
	}

	if len(form.Errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(form.Errors))
	}

	// Test with valid data
	postedData := url.Values{}

	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	r = httptest.NewRequest("POST", "/whatever", nil)

	r.PostForm = postedData

	form = NewForm(r.PostForm)

	form.Required("a", "b", "c")

	if !form.Valid() {
		t.Error("Form should have been valid")
	}

	if len(form.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(form.Errors))
	}

}

// TestForm_Check tests the Check() method of the Form type.
func TestForm_Check(t *testing.T) {

	// Test with invalid data
	form := NewForm(nil)

	form.Check(false, "password", "Password is required")

	if form.Valid() {
		t.Error("Form should not have been valid")
	}

	if len(form.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(form.Errors))
	}

}

// TestForm_ErrorGet tests the Get() method of the Form type.
func TestForm_ErrorGet(t *testing.T) {
	// Test with invalid data
	form := NewForm(nil)
	form.Check(false, "password", "Password is required")

	s := form.Errors.Get("password")

	if len(s) == 0 {
		t.Error("Expected a non-empty error string")
	}

	// Test with valid data
	s = form.Errors.Get("whatever")

	if len(s) != 0 {
		t.Error("Expected an empty error string")
	}
}

// TestForm_IsEmail tests the IsEmail() method of the Form type.
func TestForm_IsEmail(t *testing.T) {

	// Test invalid emails
	var invalidEmails = []struct {
		name  string
		input string
	}{
		{"invalid email", "me"},
		{"invalid email", "me@"},
		{"invalid email", "me@here"},
		{"invalid email", "me@here."},
		{"invalid email", "me@.com"},
		{"invalid email", ""},
	}

	for _, tt := range invalidEmails {
		f := NewForm(url.Values{"email": []string{tt.input}})
		f.Check(false, "email", "This field must be a valid email address")
		s := f.Errors.Get("email")

		if len(s) == 0 {
			t.Errorf("%s: expected an error string; got none", tt.name)
		}
	}

	// Test valid email
	var validEmail = []struct {
		name  string
		input string
	}{
		{"valid email", "me@here.com"},
	}

	for _, tt := range validEmail {
		f := NewForm(url.Values{"email": []string{tt.input}})
		f.Check(true, "email", "This field must be a valid email address")
		s := f.Errors.Get("email")

		if len(s) != 0 {
			t.Errorf("%s: expected no error string; got one", tt.name)
		}
	}

}

// TestForm_MinLength tests the MinLength() method of the Form type.
func TestForm_MinLength(t *testing.T) {

	// Test with invalid data
	var invalidLength = []struct {
		name  string
		input string
	}{
		{"invalid length", "ab"},
		{"invalid length", ""},
	}

	for _, tt := range invalidLength {
		f := NewForm(url.Values{"password": []string{tt.input}})
		f.Check(false, "password", "This field must be at least 3 characters long")
		s := f.Errors.Get("password")

		if len(s) == 0 {
			t.Errorf("%s: expected an error string; got none", tt.name)
		}
	}

	// Test with valid data
	var validLength = []struct {
		name  string
		input string
	}{
		{"valid length", "abc"},
	}

	for _, tt := range validLength {
		f := NewForm(url.Values{"password": []string{tt.input}})
		f.Check(true, "password", "This field must be at least 3 characters long")
		s := f.Errors.Get("password")

		if len(s) != 0 {
			t.Errorf("%s: expected no error string; got one", tt.name)
		}
	}
}
