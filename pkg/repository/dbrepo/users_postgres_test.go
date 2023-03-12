package dbrepo

import (
	"database/sql"
	"fmt"
	"github.com/calvarado2004/go-testing-webapp/pkg/data"
	"github.com/calvarado2004/go-testing-webapp/pkg/repository"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"os"
	"testing"
	"time"
)

//integration tests for Postgres dbrepo

var (
	host     = "localhost"
	user     = "postgres"
	password = "postgres"
	dbname   = "users_test"
	port     = "5435"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var resource *dockertest.Resource
var pool *dockertest.Pool
var testDB *sql.DB
var testRepo repository.DatabaseRepo

// TestMain is the entry point for all tests
func TestMain(m *testing.M) {

	// connect to docker; fail if docker is not running
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %v", err)
	}

	pool = p

	// set up docker options, specifying image, port, and env vars
	dockerOpts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.5",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbname,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{
					HostIP:   "0.0.0.0",
					HostPort: port,
				},
			},
		},
	}

	// get docker image
	resource, err = pool.RunWithOptions(&dockerOpts)
	if err != nil {
		log.Fatalf("Could not start resource: %v", err)
	}

	// start the docker container and wait for it to be ready
	if err = pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbname))
		if err != nil {
			log.Println("Could not connect to postgres yet")
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("Could not connect to postgres at all: %s", err)
	}

	// populate the test database
	err = createTables()
	if err != nil {
		log.Fatalf("Could not create tables: %s", err)
	}

	// initialize the repository
	testRepo = &PostgresDBRepo{DB: testDB}

	// run the tests
	code := m.Run()

	// clean up after tests
	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %v", err)
	}

	os.Exit(code)
}

// createTables creates the tables for the test database
func createTables() error {

	tableSQL, err := os.ReadFile("./testdata/users.sql")
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = testDB.Exec(string(tableSQL))
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// Test_pingDB tests the pingDB function
func Test_pingDB(t *testing.T) {
	err := testDB.Ping()
	if err != nil {
		t.Error("can't ping database")
	}
}

// TestPostgresDBRepoInsertUser tests the insertUser function
func TestPostgresDBRepoInsertUser(t *testing.T) {

	testUser := data.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := testRepo.InsertUser(testUser)
	if err != nil {
		t.Errorf("insertUser failed: %s", err)
	}

	if id != 1 {
		t.Errorf("expected id to be 1, got %d", id)
	}

}

// TestPostgresDBRepoAllUsers tests the allUsers function
func TestPostgresDBRepoAllUsers(t *testing.T) {
	users, err := testRepo.AllUsers()
	if err != nil {
		t.Errorf("allUsers failed: %s", err)
	}

	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}

	testUser2 := data.User{
		FirstName: "Jack",
		LastName:  "Smith",
		Email:     "jack@example.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, _ = testRepo.InsertUser(testUser2)

	users, err = testRepo.AllUsers()

	if err != nil {
		t.Errorf("allUsers failed: %s", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}

}

// TestPostgresDBRepoGetUser tests the getUser function
func TestPostgresDBRepoGetUser(t *testing.T) {
	user, err := testRepo.GetUser(1)
	if err != nil {
		t.Errorf("getUser failed: %s", err)
	}

	if user.ID != 1 {
		t.Errorf("expected id to be 1, got %d", user.ID)
	}

	if user.FirstName != "John" {
		t.Errorf("expected first name to be John, got %s", user.FirstName)
	}

	if user.LastName != "Doe" {
		t.Errorf("expected last name to be Doe, got %s", user.LastName)
	}

	if user.Email != "john@example.com" {
		t.Errorf("expected email to be john@example.com")
	}

	user, err = testRepo.GetUser(3)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

}

// TestPostgresDBRepoGetUserByEmail tests the getUserByEmail function
func TestPostgresDBRepoGetUserByEmail(t *testing.T) {
	user, err := testRepo.GetUserByEmail("jack@example.com")
	if err != nil {
		t.Errorf("getUserByEmail failed: %s", err)
	}

	if user.ID != 2 {
		t.Errorf("expected id to be 2, got %d", user.ID)
	}

}

// TestPostgresDBRepoUpdateUser tests the updateUser function
func TestPostgresDBRepoUpdateUser(t *testing.T) {

	user, _ := testRepo.GetUser(2)
	user.FirstName = "Jackie"
	user.LastName = "Smithy"
	user.Email = "jackie@smithy.com"

	err := testRepo.UpdateUser(*user)
	if err != nil {
		t.Errorf("updateUser failed: %s", err)
	}

	user, _ = testRepo.GetUser(2)

	if user.FirstName != "Jackie" {
		t.Errorf("expected first name to be Jackie, got %s", user.FirstName)
	}

	if user.LastName != "Smithy" {
		t.Errorf("expected last name to be Smithy, got %s", user.LastName)
	}

	if user.Email != "jackie@smithy.com" {
		t.Errorf("expected email to be jackie@smithy.com, got %s", user.Email)
	}

}

// TestPostgresDBRepoDeleteUser tests the deleteUser function
func TestPostgresDBRepoDeleteUser(t *testing.T) {
	err := testRepo.DeleteUser(1)
	if err != nil {
		t.Errorf("deleteUser failed: %s", err)
	}

	_, err = testRepo.GetUser(1)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

}

// TestPostgresDBRepoResetPassword tests the resetPassword function
func TestPostgresDBRepoResetPassword(t *testing.T) {

	err := testRepo.ResetPassword(2, "newpassword")
	if err != nil {
		t.Errorf("resetPassword failed: %s", err)
	}

	user, _ := testRepo.GetUser(2)

	matches, err := user.PasswordMatches("newpassword")
	if err != nil {
		t.Errorf("passwordMatches failed: %s", err)
	}

	if !matches {
		t.Error("expected password to match")
	}

}

// TestPostgresDBRepoInsertUserImage tests the insertUserImage function
func TestPostgresDBRepoInsertUserImage(t *testing.T) {

	testUserImage := data.UserImage{
		UserID:    2,
		FileName:  "test.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := testRepo.InsertUserImage(testUserImage)
	if err != nil {
		t.Errorf("insertUserImage failed: %v", err)
	}

	if id != 1 {
		t.Errorf("expected id to be 1, got %d", id)
	}

}
