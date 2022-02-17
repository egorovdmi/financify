// Small web framework
package web

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/uuid"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// KeyValues is how request values are stored/retrieved.
const KeyValues ctxKey = 1

// Values represent state for each request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

type Handler func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error

type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
}

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
		v := Values{
			TraceID: uuid.New().String(),
			Now:     time.Now(),
		}
		ctx := context.WithValue(r.Context(), KeyValues, &v)

		if err := handler(ctx, rw, r); err != nil {
			a.SignalShutdown()
			return
		}

		// BOILERPLATE
	}

	a.ContextMux.Handle(method, path, h)
}
