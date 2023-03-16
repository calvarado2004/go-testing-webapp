package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/calvarado2004/go-testing-webapp/pkg/repository"
	"github.com/calvarado2004/go-testing-webapp/pkg/repository/dbrepo"
	"log"
	"net/http"
	"os"
)

const port = 8090

type application struct {
	DSN       string
	DB        repository.DatabaseRepo
	Domain    string
	JWTSecret string
}

func main() {
	var app application

	flag.StringVar(&app.Domain, "domain", "example.com", "Domain name of the application")
	flag.StringVar(&app.DSN, "dsn", os.Getenv("DSN"), "Postgres DSN")
	flag.StringVar(&app.JWTSecret, "jwt-secret", os.Getenv("JWT_SECRET"), "JWT Secret")
	flag.Parse()

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}

	defer func(conn *sql.DB) {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conn)

	app.DB = &dbrepo.PostgresDBRepo{DB: conn}

	log.Printf("Starting server on port %d", port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())
	if err != nil {
		log.Fatal(err)
	}

}
