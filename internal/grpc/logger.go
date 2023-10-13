package grpc

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

// LoggerUnaryServerInterceptor returns a new unary server interceptor logging.
func LoggerUnaryServerInterceptor(opts ...LoggerOption) grpc.UnaryServerInterceptor {
	o := evaluateLoggerOptions(opts)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ any, err error) {
		le := &logEntry{o.l, o.concise, o.sensitive, o.leak}
		ctx = context.WithValue(ctx, ContextKeyLogEntry, le)

		t := time.Now()
		defer func() {
			le.error(err)
			le.log(ctx, info.FullMethod, time.Since(t), err)
		}()

		return handler(ctx, req)
	}
}

// LogEntryAttr helper func for setting slog.Attr to the logEntry.
func LogEntryAttr(ctx context.Context, a ...any /* slog.Attr */) {
	if entry, ok := ctx.Value(ContextKeyLogEntry).(*logEntry); ok {
		entry.l = entry.l.With(a...)
	}
}

// LogEntryError helper func for adding error message to the log.
// NOT NEEDED since we have err in the middleware.
// func LogEntryError(ctx context.Context, msg string) {
// 	if entry, ok := ctx.Value(ContextKeyLogEntry).(*logEntry); ok {
// 		entry.l = entry.l.With(slog.String("error", msg))
// 	}
// }

type logEntry struct {
	l         *slog.Logger
	concise   bool
	sensitive map[string]struct{}
	leak      bool
}

func (le *logEntry) error(err error) {
	if err != nil && err != io.EOF {
		le.l = le.l.With(slog.String("error", err.Error()))
	}
}

func (le *logEntry) log(ctx context.Context, method string, d time.Duration, err error) {
	reqID, ok := ctx.Value(ContextKeyRequestID).(string)
	if ok {
		le.l = le.l.With(slog.String("id", reqID))
	}
	code := status.Code(err)
	le.l = le.l.With(
		slog.String("method", method),
		slog.Group("status",
			slog.Int("code", int(code)),
			slog.String("msg", code.String()),
		),
	)

	if in, ok := metadata.FromIncomingContext(ctx); ok && !le.concise {
		incomingAttr := grpcMetadataAttrs(in, le.leak, le.sensitive)
		le.l = le.l.With(slog.Group("incoming", incomingAttr...))
	}
	if out, ok := metadata.FromOutgoingContext(ctx); ok && !le.concise {
		outgoingAttr := grpcMetadataAttrs(out, le.leak, le.sensitive)
		le.l = le.l.With(slog.Group("outgoing", outgoingAttr...))
	}

	le.l.LogAttrs(
		ctx,
		toLogLevel(code),
		fmt.Sprintf("%d %s", code, code.String()),
		slog.Duration("duration", d),
	)
}

func grpcMetadataAttrs(metadata metadata.MD, leak bool, sensitive map[string]struct{}) []any {
	metadatAttr := make([]any, 0, len(metadata)) // []slog.Attr

	for k, v := range metadata {
		k = strings.ToLower(k)
		_, ok := sensitive[k]

		switch {
		case ok && !leak: // filtering sensitive headers
			continue
		case len(v) == 0:
			continue
		case len(v) == 1:
			metadatAttr = append(metadatAttr, slog.String(k, v[0]))
		default:
			metadatAttr = append(metadatAttr, slog.String(k, fmt.Sprintf("[%s]", strings.Join(v, "], ["))))
		}
	}

	return metadatAttr
}

func toLogLevel(code codes.Code) slog.Level {
	switch code {
	case codes.OK:
		return slog.LevelInfo
	case codes.NotFound, codes.Canceled, codes.AlreadyExists, codes.InvalidArgument, codes.Unauthenticated, codes.PermissionDenied, codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted, codes.OutOfRange:
		return slog.LevelWarn
	case codes.DeadlineExceeded, codes.Internal, codes.Unavailable, codes.Unknown, codes.Unimplemented, codes.DataLoss:
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
			o.sensitive = make(map[string]struct{}, 1)
		} else {
			o.sensitive = make(map[string]struct{}, len(s))
			for k := range s {
				o.sensitive[k] = struct{}{}
			}
		}
		s["authorization"] = struct{}{}
	}
}

// only for dev purposes
func WithLeak(leak bool) LoggerOption {
	return func(o *loggerOptions) {
		o.leak = leak
	}
}
