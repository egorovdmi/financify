package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/egorovdmi/financify/business/auth"
	"github.com/egorovdmi/financify/business/mid"
	"github.com/egorovdmi/financify/foundation/web"
)

func API(build string, shutdown chan os.Signal, log *log.Logger, a *auth.Auth) *web.App {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	check := check{
		log: log,
	}
	app.Handle(http.MethodGet, "/readiness", check.readiness)
	app.Handle(http.MethodGet, "/liveness", check.readiness)

	return app
}
