package handlers

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"

	"github.com/egorovdmi/financify/foundation/web"
)

type check struct {
	log *log.Logger
}

func (c check) readiness(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	// Simulation of an error. 50% of requests will be returning an error
	if n := rand.Intn(100); n%2 == 0 {
		return errors.New("something went wrong")
	}

	status := struct {
		Status string
	}{
		Status: "OK",
	}
	return web.Respond(ctx, rw, status, http.StatusOK)
}
