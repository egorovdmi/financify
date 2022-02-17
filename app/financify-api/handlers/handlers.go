package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/dimfeld/httptreemux/v5"
)

func API(build string, shutdown chan os.Signal, log *log.Logger) *httptreemux.ContextMux {
	router := httptreemux.NewContextMux()

	check := check{
		log: log,
	}
	router.Handle(http.MethodGet, "/readiness", check.readiness)

	return router
}
