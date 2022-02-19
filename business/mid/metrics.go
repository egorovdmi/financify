package mid

import (
	"context"
	"expvar"
	"net/http"
	"runtime"

	"github.com/egorovdmi/financify/foundation/web"
)

var m = struct {
	err *expvar.Int
	req *expvar.Int
	gr  *expvar.Int
}{
	err: expvar.NewInt("errors"),
	req: expvar.NewInt("requests"),
	gr:  expvar.NewInt("goroutines"),
}

func Metrics() web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
			err := handler(ctx, rw, r)

			m.req.Add(1)

			if err != nil {
				m.err.Add(1)
			}

			// each 10 requests we update goroutines counter
			if m.req.Value()%10 == 0 {
				m.gr.Set(int64(runtime.NumGoroutine()))
			}

			return err
		}

		return h
	}

	return m
}
