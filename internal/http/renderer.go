package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	e "go-misc/internal/errors"
)

func DecodeRequest(req *http.Request, v interface{}) error {
	err := json.NewDecoder(req.Body).Decode(v)
	if err != nil && err != io.EOF {
		return fmt.Errorf("decoding request: %w", e.Wrap(e.CodeInvalidArgument, err))
	}
	return nil
}

func EncodeResponse(w http.ResponseWriter, v interface{}) error {
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		return fmt.Errorf("encoding response: %v%w", err, e.ErrInternal)
	}
	return nil
}

func EncodeError(ctx context.Context, w http.ResponseWriter, err error) {
	if err == nil {
		panic("error: err cannot be nil")
	}
	code := http.StatusInternalServerError
	// log the error by adding the msg to the logEntry
	LogEntryError(ctx, err)

	var ierr *e.Error
	if errors.As(err, &ierr) {
		code = e.HttpStatus(ierr.Code())
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(struct {
			Err string `json:"error,omitempty"`
		}{Err: ierr.Error()})
		return
	}

	var ierrs *e.Errors
	if errors.As(err, &ierrs) {
		code = e.HttpStatus(ierrs.Code())
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(struct {
			Errs []string `json:"errors,omitempty"`
		}{Errs: ierrs.Errors()})
		return
	}
}
