package handlers

import (
	"context"
	"net/http"

	"github.com/egorovdmi/financify/business/auth"
	"github.com/egorovdmi/financify/business/data/user"
	"github.com/egorovdmi/financify/foundation/web"
	"github.com/pkg/errors"
)

type userGroup struct {
	repo user.UserRepository
	auth *auth.Auth
}

func (ug userGroup) query(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	users, err := ug.repo.Query(ctx, v.TraceID)
	if err != nil {
		return errors.Wrap(err, "unable to query for users")
	}

	return web.Respond(ctx, rw, users, http.StatusOK)
}

func (ug userGroup) queryByID(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	usr, err := ug.repo.QueryByID(ctx, v.TraceID, claims, web.Param(r, "id"))
	if err != nil {
		switch err {
		case user.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case user.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case user.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s", web.Param(r, "id"))
		}
	}

	return web.Respond(ctx, rw, &usr, http.StatusOK)
}

func (ug userGroup) create(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var nu user.NewUser
	if err := web.Decode(r, &nu); err != nil {
		return errors.Wrap(err, "unable to decode payload")
	}

	usr, err := ug.repo.Create(ctx, v.TraceID, nu, v.Now)
	if err != nil {
		return errors.Wrapf(err, "User: %+v", &nu)
	}

	return web.Respond(ctx, rw, &usr, http.StatusCreated)
}

func (ug userGroup) update(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	var uu user.UpdateUser
	if err := web.Decode(r, &uu); err != nil {
		return errors.Wrap(err, "unable to decode payload")
	}

	if err := ug.repo.Update(ctx, v.TraceID, claims, web.Param(r, "id"), uu, v.Now); err != nil {
		switch err {
		case user.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case user.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case user.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s; User: %+v", web.Param(r, "id"), &uu)
		}
	}

	return web.Respond(ctx, rw, nil, http.StatusNoContent)
}

func (ug userGroup) delete(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	if err := ug.repo.Delete(ctx, v.TraceID, web.Param(r, "id")); err != nil {
		switch err {
		case user.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case user.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case user.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s", web.Param(r, "id"))
		}
	}

	return web.Respond(ctx, rw, nil, http.StatusNoContent)
}

func (ug userGroup) token(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := ug.repo.Authenticate(ctx, v.TraceID, email, pass, v.Now)
	if err != nil {
		switch err {
		case user.ErrAuthenticationFailure:
			return web.NewRequestError(err, http.StatusUnauthorized)
		default:
			return errors.Wrap(err, "authenticating")
		}
	}

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = ug.auth.GenerateToken(web.Param(r, "kid"), claims)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	return web.Respond(ctx, rw, tkn, http.StatusOK)
}
