package main

import (
	"github.com/calvarado2004/go-testing-webapp/pkg/repository/dbrepo"
	"os"
	"testing"
)

var app application

func TestMain(m *testing.M) {

	app.DB = &dbrepo.TestDBRepo{}

	app.Domain = "example.com"

	app.JWTSecret = "verysecret"

	os.Exit(m.Run())

}
