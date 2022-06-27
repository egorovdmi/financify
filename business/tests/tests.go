package tests

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/egorovdmi/financify/business/auth"
	"github.com/egorovdmi/financify/business/data/dbschema"
	"github.com/egorovdmi/financify/business/data/user"
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

type Test struct {
	TraceID string
	DB      *sqlx.DB
	Log     *log.Logger
	Auth    *auth.Auth
	KID     string

	t       *testing.T
	cleanup func()
}

func NewIntegration(t *testing.T) *Test {
	log, db, teardown := NewUnit(t)
	ctx := Context()

	if err := dbschema.Seed(ctx, db); err != nil {
		t.Fatal(err)
	}

	// Generate RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	// Build an authenticator
	const keyID = "7a7fb378-d885-43ad-aa25-a0b33bca287f"
	lookup := func(kid string) (*rsa.PublicKey, error) {
		switch kid {
		case keyID:
			return &privateKey.PublicKey, nil
		}
		return nil, fmt.Errorf("no public key found for the specified kid: %s", kid)
	}

	auth, err := auth.New("RS256", lookup, auth.Keys{keyID: privateKey})
	if err != nil {
		t.Fatal(err)
	}

	test := Test{
		TraceID: "00000000-0000-0000-0000-000000000000",
		DB:      db,
		Log:     log,
		Auth:    auth,
		KID:     keyID,
		t:       t,
		cleanup: teardown,
	}

	return &test
}

func (test *Test) Teardown() {
	test.cleanup()
}

func (test *Test) Token(kid string, email string, pass string) string {
	userRepo := user.NewUserRepository(test.Log, test.DB)
	claims, err := userRepo.Authenticate(context.Background(), test.TraceID, email, pass, time.Now())
	if err != nil {
		test.t.Fatal(err)
	}

	token, err := test.Auth.GenerateToken(test.KID, claims)
	if err != nil {
		test.t.Fatal(err)
	}

	return token
}
