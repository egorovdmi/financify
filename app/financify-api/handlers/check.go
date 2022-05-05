package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/egorovdmi/financify/foundation/web"
)

type check struct {
	log *log.Logger
}

func (c check) readiness(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	status := struct {
		Status string
	}{
		Status: "OK",
	}
	return web.Respond(ctx, rw, status, http.StatusOK)
}
