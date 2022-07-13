package mid

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/egorovdmi/financify/foundation/web"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
)

func Panics(log *log.Logger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, rw http.ResponseWriter, r *http.Request) (err error) {
			currentSpan := trace.SpanFromContext(ctx)
			ctx, span := currentSpan.TracerProvider().Tracer("").Start(ctx, "business.mid.panics")
			defer span.End()

			// Extracting TraceID from the context
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("missing KeyValues in the context")
			}

			defer func() {
				if r := recover(); r != nil {
					err = errors.Errorf("panic: %v", r)

					log.Printf("%s : FATAL     : %s", v.TraceID, debug.Stack())
				}
			}()

			return handler(ctx, rw, r)
		}

		return h
	}

	return m
}
