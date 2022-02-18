package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

func Respond(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error {
	// Set the status code for the request logger middleware.
	v, ok := ctx.Value(KeyValues).(*Values)
	if !ok {
		return NewShutdownError("tracing value missing from context")
	}
	v.StatusCode = statusCode

	// If there is nothing to marshal then set status code and return
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Send the result code to the response
	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}

func RespondError(ctx context.Context, w http.ResponseWriter, err error) error {
	// If the error was of the type *Error, the handler has
	// a specific status code and error to return.
	if webErr, ok := errors.Cause(err).(*Error); ok {
		er := ErrorResponse{
			Error:  webErr.Err.Error(),
			Fields: webErr.Fields,
		}

		if err := Respond(ctx, w, er, webErr.Status); err != nil {
			return err
		}

		return nil
	}

	// If not, the handler sent any regular error value so use 500.
	er := ErrorResponse{
		Error: http.StatusText(http.StatusInternalServerError),
	}

	if err := Respond(ctx, w, er, http.StatusInternalServerError); err != nil {
		return err
	}

	return nil
}
