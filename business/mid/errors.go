package mid

import (
	"context"
	"log"
	"net/http"

	"github.com/egorovdmi/financify/foundation/web"
)

func Errors(log *log.Logger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {

			// Extracting TraceID from the context
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("missing KeyValues in the context")
			}

			// execute core handler
			if err := handler(ctx, rw, r); err != nil {
				log.Printf("%s : ERROR     : %+v", v.TraceID, err)

				if err := web.RespondError(ctx, rw, err); err != nil {
					return err
				}

				if web.IsShutdown(err) {
					return err
				}
			}

			return nil
		}

		return h
	}

	return m
}
