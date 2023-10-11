# go-misc
A collection of Go codes and how to use them.

## hello
A little server that says hello... and uses all the Go codes mentioned below.

[cmd/main.go](./cmd/main.go) to check how they are glued together.

### `http` localhost:8000/say/hello
```json
// response
{
    "message": "Hello, Bonjour, Salama"
}
```

### `grpc` localhost:8001/say
```json
// request
{
    "id": "hello"
}

// response
{
    "message": "Hello, Bonjour, Salama"
}
```

## logger
A request logger middleware/interceptor using `log/slog`.
### `configuration` 
[logger/logger.go](./internal/logger/logger.go)
```go
// use case
l := logger.NewLogger(
    logger.WithFormat("json"),
    logger.WithLevel(slog.LevelDebug),
    logger.WithServiceName("hello"),
    logger.WithTags(map[string]string{
        "version": "v1.0-81aa4244d9fc8076a",
        "env":     "dev",
    }),
)
```
### `http` 
[http/logger.go](./internal/http/logger.go)
```go
// use case
r := mux.NewRouter()
r.Use(
    httpmw.ContentType,
    httpmw.RequestID, //before logger
    httpmw.Logger(
        httpmw.WithLogger(httpl),
        httpmw.WithConcise(true),
        httpmw.WithLeak(false),
        httpmw.WithSensitive(map[string]struct{}{
            "insecure":       {},
            "very-insercure": {},
        }),
    ),
    httpmw.Recover, // after Logger
)
```

### `grpc` 
[http/logger.go](./internal/grpc/logger.go)
```go
// use case
grpcsrv := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        grpcmw.RequestIDUnaryServerInterceptor, // before logger
        grpcmw.LoggerUnaryServerInterceptor(
            grpcmw.WithLogger(grpcl),
            grpcmw.WithConcise(true),
            grpcmw.WithLeak(false),
            grpcmw.WithSensitive(map[string]struct{}{
                "insecure":       {},
                "very-insercure": {},
            }),
        ),
        grpcmw.RecoverUnaryServerInterceptor, // after error
    ),
)
```

## requestID
A middlerware/interceptor for setting a request ID.

[http/request_id.go](./internal/http/reuqest_id.go)

[grpc/request_id.go](./internal/grpc/reuqest_id.go)

## recover
A middlerware/interceptor for recovering from a panic.

[http/recover.go](./internal/http/recover.go)

[grpc/recover.go](./internal/grpc/recover.go)

## other
[http/wrap_writer.go](./internal/http/wrap_writer.go)

for wrapping the an `http.ResponseWriter` and add `code` and `size` field.

[http/renderer.go](./internal/http/renderer.go)

helper functions for decoding requests, encoding response and errors.

## validator
A simple warpper arround `go-playground/validator` `Struct()` method for a cusotm error and error messages.

[validator/validator.go](./internal/validator/validator.go)

## errors
An error type with a `code` that can be translated to an `http.Status` or `grpc.Status`.

[errors/errors.go](./internal/errors/errors.go)

## cache
An in-memory cache that uses `Allegro/BigCache`.

[cache/cache.go](./internal/cache/cache.go)