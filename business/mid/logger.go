package mid

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/egorovdmi/financify/foundation/web"
)

func Logger(log *log.Logger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {

			// Extracting TraceID from the context
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("missing KeyValues in the context")
			}

			log.Printf("%s : started   : %s %s -> %s", v.TraceID, r.Method, r.URL.Path,
				r.RemoteAddr)

			// wrapped core handler
			err := handler(ctx, rw, r)

			log.Printf("%s : completed : %s %s -> %s (%d) (%s)", v.TraceID, r.Method, r.URL.Path,
				r.RemoteAddr, v.StatusCode, time.Since(v.Now))

			return err
		}

		return h
	}

	return m
}
