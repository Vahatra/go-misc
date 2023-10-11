package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type loggerOptions struct {
	l         *slog.Logger
	concise   bool                // detailed or concise logs
	sensitive map[string]struct{} // a set for storing fields that should not be logged
	leak      bool                // ignore "sensitive" and log everything
}

type LoggerOption func(*loggerOptions)

func evaluateLoggerOptions(opts []LoggerOption) *loggerOptions {
	opt := &loggerOptions{
		l:         slog.Default(),
		concise:   false,
		sensitive: nil,
		leak:      false,
	}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

func Logger(opts ...LoggerOption) func(next http.Handler) http.Handler {
	o := evaluateLoggerOptions(opts)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := &wWriter{w, false, http.StatusOK, 0}
			le := &logEntry{o.l, o.concise, o.sensitive, o.leak}
			r = r.WithContext(context.WithValue(r.Context(), ContextKeyLogEntry, le))

			t := time.Now()
			defer func() {
				le.req(r)
				le.resp(ww)
				le.log(ww.code, time.Since(t))
			}()

			next.ServeHTTP(reimplement(ww), r)
		})
	}
}

// LogEntryAttr helper func for setting slog.Attr to the logEntry.
func LogEntryAttr(ctx context.Context, attr ...any /* slog.Attr */) {
	if entry, ok := ctx.Value(ContextKeyLogEntry).(*logEntry); ok {
		entry.l = entry.l.With(attr...)
	}
}

// LogEntryError helper func for adding error message to the log.
func LogEntryError(ctx context.Context, msg string) {
	if entry, ok := ctx.Value(ContextKeyLogEntry).(*logEntry); ok {
		entry.l = entry.l.With(slog.String("error", msg))
	}
}

type logEntry struct {
	l         *slog.Logger
	concise   bool
	sensitive map[string]struct{}
	leak      bool
}

func (le *logEntry) req(r *http.Request) {
	reqID, ok := r.Context().Value(ContextKeyRequestID).(string)
	if ok {
		le.l = le.l.With(slog.String("id", reqID))
	}

	requestAttr := make([]any, 0, 7) // slog.Attr
	requestAttr = append(requestAttr,
		slog.String("uri", r.RequestURI),
		slog.String("method", r.Method),
	)

	if !le.concise {
		requestAttr = append(requestAttr,
			slog.String("host", r.Host),
			slog.String("path", r.URL.Path),
			slog.String("proto", r.Proto),
			slog.String("remote", r.RemoteAddr),
			httpHeaderAttrs(r.Header, le.leak, le.sensitive),
		)
	}
	le.l = le.l.With(slog.Group("request", requestAttr...))
}

func (le *logEntry) resp(w *wWriter) {
	responseAttr := make([]any, 0, 3) // slog.Attr
	responseAttr = append(responseAttr,
		slog.Int("size", w.size),
		slog.Group("status",
			slog.Int("code", w.code),
			slog.String("msg", http.StatusText(w.code)),
		))
	if !le.concise {
		responseAttr = append(responseAttr, httpHeaderAttrs(w.Header(), le.leak, le.sensitive))
	}
	le.l = le.l.With(slog.Group("response", responseAttr...))
}

func (le *logEntry) log(code int, d time.Duration) {
	msg := fmt.Sprintf("%d %s", code, http.StatusText(code))
	le.l.LogAttrs(
		nil,
		toLogLevel(code),
		msg,
		slog.Duration("duration", d),
	)
}

func httpHeaderAttrs(header http.Header, leak bool, sensitive map[string]struct{}) slog.Attr {
	hearderAttr := make([]any, 0, len(header)) // []slog.Attr

	for k, v := range header {
		k = strings.ToLower(k)
		_, ok := sensitive[k]

		switch {
		case ok && !leak: // filtering sensitive headers
			continue
		case len(v) == 0:
			continue
		case len(v) == 1:
			hearderAttr = append(hearderAttr, slog.String(k, v[0]))
		default:
			hearderAttr = append(hearderAttr, slog.String(k, fmt.Sprintf("[%s]", strings.Join(v, "], ["))))
		}
	}

	return slog.Group("headers", hearderAttr...)
}

func toLogLevel(status int) slog.Level {
	switch {
	case status <= 0:
		return slog.LevelWarn
	case status < 400: // for codes in 100s, 200s, 300s
		return slog.LevelInfo
	case status >= 400 && status < 500:
		return slog.LevelWarn
	case status >= 500:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithLoggerLogger is a functional option to use another *slog.Logger
func WithLogger(l *slog.Logger) LoggerOption {
	return func(o *loggerOptions) {
		o.l = l
	}
}

func WithConcise(concise bool) LoggerOption {
	return func(o *loggerOptions) {
		o.concise = concise
	}
}

func WithSensitive(s map[string]struct{}) LoggerOption {
	return func(o *loggerOptions) {
		// https://github.com/uber-go/guide/blob/master/style.md#copy-slices-and-maps-at-boundaries
		if s == nil {
			o.sensitive = make(map[string]struct{}, 3)
		} else {
			o.sensitive = make(map[string]struct{}, len(s))
			for k := range s {
				o.sensitive[k] = struct{}{}
			}
		}
		s["authorization"] = struct{}{}
		s["cookie"] = struct{}{}
		s["set-cookie"] = struct{}{}
	}
}

// only for dev purposes
func WithLeak(leakSensitiveData bool) LoggerOption {
	return func(o *loggerOptions) {
		o.leak = leakSensitiveData
	}
}
