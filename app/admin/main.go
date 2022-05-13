package main

import (
	"context"
	"log"

	"github.com/egorovdmi/financify/business/data/dbschema"
	"github.com/egorovdmi/financify/foundation/database"
)

func main() {
	db, err := database.Open(database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       "localhost",
		Name:       "financify",
		DisableTLS: true,
	})

	if err != nil {
		log.Fatalf("opennig database error, %v", err)
	}

	defer db.Close()

	err = dbschema.Migrate(context.Background(), db)
	if err != nil {
		log.Fatalf("database migration error, %v", err)
	}

	log.Println("Database migration complete")
}
