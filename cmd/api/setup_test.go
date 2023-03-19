package main

import (
	"github.com/calvarado2004/go-testing-webapp/pkg/repository/dbrepo"
	"os"
	"testing"
)

var app application

var expiredToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiYXVkIjoiZXhhbXBsZS5jb20iLCJleHAiOjE2Nzg4OTQzMTYsImlzcyI6ImV4YW1wbGUuY29tIiwibmFtZSI6IkpvaG4gRG9lIiwic3ViIjoiMSJ9.ytNHyK5mq9cB1mwP9hccHHx77Qon5iHmECTUIVfq620"

func TestMain(m *testing.M) {

	app.DB = &dbrepo.TestDBRepo{}

	app.Domain = "example.com"

	app.JWTSecret = "2dce505d96a53c5768052ee90f3df2055657518dad489160df9913f66042e160"

	os.Exit(m.Run())

}
