package main

// openDB returns a sql.DB connection pool for the named data source.
import (
	"database/sql"
	"log"
)

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// connectToDB returns a sql.DB connection pool for the named data source.
func (app *application) connectToDB() (*sql.DB, error) {

	connection, err := openDB(app.DSN)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to Postgres!")

	return connection, nil

}
