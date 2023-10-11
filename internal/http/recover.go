package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if re := recover(); re != nil {
				stack := string(debug.Stack())
				LogEntryAttr(
					r.Context(),
					slog.String("error", fmt.Sprintf("panic caught: %v", re)),
					slog.String("stack", stack),
				)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
