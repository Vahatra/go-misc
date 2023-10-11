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
	var msg string
	var ie *e.Error
	if errors.As(err, &ie) {
		code = ie.ToHTTPStatus()
		if ie.Error() != "" {
			msg = ie.Error()
		}
	}
	w.WriteHeader(code)
	if msg != "" {
		json.NewEncoder(w).Encode(struct {
			Err string `json:"error,omitempty"`
		}{Err: msg})
	}
	// log the error by adding the msg to the logEntry
	LogEntryError(ctx, err.Error())
}
