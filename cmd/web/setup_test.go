package main

import (
	"github.com/calvarado2004/go-testing/pkg/repository/dbrepo"
	"os"
	"testing"
)

var app application

// TestMain is the entry point for all tests.
func TestMain(m *testing.M) {

	pathToTemplates = "./../../templates/"

	app.Session = getSession()

	app.DB = &dbrepo.TestDBRepo{}

	os.Exit(m.Run())
}
