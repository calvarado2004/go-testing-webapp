package main

import (
	"database/sql"
	"encoding/gob"
	"flag"
	"github.com/alexedwards/scs/v2"
	"github.com/calvarado2004/go-testing-webapp/pkg/data"
	"github.com/calvarado2004/go-testing-webapp/pkg/repository"
	"github.com/calvarado2004/go-testing-webapp/pkg/repository/dbrepo"
	"log"
	"net/http"
	"os"
)

type application struct {
	DB      repository.DatabaseRepo
	DSN     string
	Session *scs.SessionManager
}

//DSN="host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5"; JWT_SECRET="verysecret" go run ./cmd/web

func main() {

	// register the user model with gob
	gob.Register(data.User{})

	// set up an app config
	app := application{}

	// read the DSN from the command line or environment variable
	flag.StringVar(&app.DSN, "dsn", os.Getenv("DSN"), "Postgres DSN")
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

	// get a session manager
	app.Session = getSession()

	// get application routes
	mux := app.routes()

	// print out a message
	log.Println("Starting server on :8080")

	// start the server
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
