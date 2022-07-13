package mid

import (
	"context"
	"net/http"
	"strings"

	"github.com/egorovdmi/financify/business/auth"
	"github.com/egorovdmi/financify/foundation/web"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
)

var ErrForbidden = web.NewRequestError(
	errors.New("you are not authorized for that action"),
	http.StatusForbidden)

func Authenticate(a *auth.Auth) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, rw http.ResponseWriter, r *http.Request) (err error) {
			currentSpan := trace.SpanFromContext(ctx)
			ctx, span := currentSpan.TracerProvider().Tracer("").Start(ctx, "business.mid.authenticate")
			defer span.End()

			// Parse the authorization header.
			// Expected: `Bearer <token>`.
			authHeaderValue := r.Header.Get("Authorization")
			parts := strings.Split(authHeaderValue, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.New("expected authorization header format: Bearer <token>")
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			// Validate the token is signed by us.
			claims, err := a.ValidateToken(parts[1])
			if err != nil {
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			// Add claims to the context so they can be retrieved later.
			ctx = context.WithValue(ctx, auth.Key, claims)

			return handler(ctx, rw, r)
		}

		return h
	}

	return m
}

func Authorize(roles ...string) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, rw http.ResponseWriter, r *http.Request) (err error) {
			currentSpan := trace.SpanFromContext(ctx)
			ctx, span := currentSpan.TracerProvider().Tracer("").Start(ctx, "business.mid.authorize")
			defer span.End()

			// Extracting TraceID from the context
			claims, ok := ctx.Value(auth.Key).(auth.Claims)
			if !ok {
				return errors.New("missing claims in the context: Authorize called without/before Authenticate middleware")
			}

			if !claims.Authorize(roles...) {
				return ErrForbidden
			}

			return handler(ctx, rw, r)
		}

		return h
	}

	return m
}
