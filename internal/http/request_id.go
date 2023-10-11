package http

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey int

const (
	// generated uuid.
	ContextKeyRequestID contextKey = iota

	// log entry.
	ContextKeyLogEntry
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := uuid.New()
		ctx = context.WithValue(ctx, ContextKeyRequestID, id.String())
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
