package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/egorovdmi/financify/foundation/web"
)

func API(build string, shutdown chan os.Signal, log *log.Logger) *web.App {
	app := web.NewApp(shutdown)

	check := check{
		log: log,
	}
	app.Handle(http.MethodGet, "/readiness", check.readiness)

	return app
}
