package tests

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/egorovdmi/financify/business/data/dbschema"
	"github.com/egorovdmi/financify/foundation/database"
	"github.com/egorovdmi/financify/foundation/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Success and failure markers.
const (
	Success = "\033[32m\u2713\033[0m"
	Failed  = "\033[31m\u2717\033[0m"
)

// Configuration for runnung tests.
var (
	dbImage = "postgres:14-alpine"
	dbPort  = "5432"
	dbArgs  = []string{"-e", "POSTGRES_PASSWORD=postgres"}
	AdminID = "5cf37266-3473-4006-984f-9325122678b7"
	UserID  = "45b5fbd3-755f-4379-8f07-a58d4a30fa2f"
)

func NewUnit(t *testing.T) (*log.Logger, *sqlx.DB, func()) {
	c := startContainer(t, dbImage, dbPort, dbArgs...)

	cfg := database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.Host,
		Name:       "postgres",
		DisableTLS: true,
	}
	db, err := database.Open(cfg)
	if err != nil {
		t.Fatalf("opening database connection: %s", err)
	}

	t.Log("waiting for database to be ready ...")

	// Wait for the database to be ready.
	var pingError error
	maxAttemts := 20
	for attempts := 1; attempts < maxAttemts; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
	}

	if pingError != nil {
		dumpContainerLogs(t, c.ID)
		stopContainer(t, c.ID)
		t.Fatalf("database never ready: %s", pingError)
	}

	if err := dbschema.Migrate(context.Background(), db); err != nil {
		stopContainer(t, c.ID)
		t.Fatalf("mimgrating error: %s", err)
	}

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown := func() {
		t.Helper()
		db.Close()
		stopContainer(t, c.ID)
	}

	log := log.New(os.Stdout, "TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	return log, db, teardown
}

// Context returns an app level context for testing.
func Context() context.Context {
	values := web.Values{
		TraceID: uuid.New().String(),
		Now:     time.Now(),
	}

	return context.WithValue(context.Background(), web.KeyValues, &values)
}

// StringPointer is a helper method for take a pointer of a string. Useful for tests.
func StringPointer(s string) *string {
	return &s
}

// IntPointer is a helper method for take a pointer of an int. Useful for tests.
func IntPointer(i int) *int {
	return &i
}
