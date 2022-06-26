package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/egorovdmi/financify/business/auth"
	"github.com/egorovdmi/financify/business/data/user"
	"github.com/egorovdmi/financify/business/mid"
	"github.com/egorovdmi/financify/foundation/web"
	"github.com/jmoiron/sqlx"
)

func API(build string, shutdown chan os.Signal, log *log.Logger, a *auth.Auth, db *sqlx.DB) *web.App {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	check := checkGroup{
		build: build,
		db:    db,
	}
	app.Handle(http.MethodGet, "/readiness", check.readiness)
	app.Handle(http.MethodGet, "/liveness", check.liveness)

	ug := userGroup{
		repo: user.NewUserRepository(log, db),
		auth: a,
	}

	app.Handle(http.MethodGet, "/v1/token/:kid", ug.token)
	app.Handle(http.MethodGet, "/v1/users", ug.query, mid.Authenticate(a))
	app.Handle(http.MethodGet, "/v1/users/:id", ug.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPost, "/v1/users", ug.create, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/v1/users/:id", ug.update, mid.Authenticate(a))
	app.Handle(http.MethodDelete, "/v1/users/:id", ug.delete, mid.Authenticate(a))

	return app
}
