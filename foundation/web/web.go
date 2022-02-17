package web

import (
	"context"
	"net/http"
	"os"
	"syscall"

	"github.com/dimfeld/httptreemux/v5"
)

type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
}

type Handler func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error

func NewApp(shutdown chan os.Signal) *App {
	app := App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
	}

	return &app
}

func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

func (a *App) Handle(method string, path string, handler Handler) {
	h := func(rw http.ResponseWriter, r *http.Request) {

		// BOILERPLATE

		if err := handler(r.Context(), rw, r); err != nil {
			a.SignalShutdown()
			return
		}

		// BOILERPLATE
	}

	a.ContextMux.Handle(method, path, h)
}
