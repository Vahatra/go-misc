package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go-misc/internal/cache"
	grpcmw "go-misc/internal/grpc"
	"go-misc/internal/grpc/pb"
	"go-misc/internal/hello"
	hellogrpc "go-misc/internal/hello/grpc"
	hellohttp "go-misc/internal/hello/http"
	helloinmem "go-misc/internal/hello/inmem"
	httpmw "go-misc/internal/http"
	"go-misc/internal/logger"
	"go-misc/internal/validator"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	signalCtx, signalCancel := context.WithCancel(context.Background())

	l := logger.NewLogger(
		logger.WithFormat("json"),
		logger.WithLevel(slog.LevelDebug),
		logger.WithServiceName("wallet"),
		// logger.WithTags(map[string]string{
		// 	"version": "v1.0-81aa4244d9fc8076a",
		// 	"env":     "dev",
		// }),
	)
	v := validator.NewValidator()
	c, err := cache.NewCache(signalCtx)
	if err != nil {
		l.Error("error creating cache", "err", err.Error())
		os.Exit(1)
	}

	helloRepo := helloinmem.NewRepository()
	helloService := hello.NewService(l, v, c, helloRepo)

	var wg sync.WaitGroup
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		<-signals
		l.Debug("Interrupt signal.")

		signalCancel()
	}()

	// HTTP
	wg.Add(1)
	go func() {
		httpl := l.With(slog.String("transport", "http"))
		r := mux.NewRouter()
		r.Use(
			httpmw.ContentType,
			httpmw.RequestID, //before logger
			httpmw.Logger(
				httpmw.WithLogger(httpl),
				httpmw.WithConcise(true),
				httpmw.WithLeak(false),
			// httpmw.WithSensitive(map[string]struct{}{
			// 	"insecure":       {},
			// 	"very-insercure": {},
			// }),
			),
			httpmw.Recover, // after Logger
		)

		helloHandler := hellohttp.NewHandler(httpl, helloService)
		helloHandler.Register(r.PathPrefix("/v1").Subrouter())

		httpsrv := &http.Server{
			Handler: r,
			Addr:    "0.0.0.0:8000",
		}

		go func() {
			defer wg.Done()
			<-signalCtx.Done() // Wait for the context to be done

			s, c := context.WithTimeout(context.Background(), 10*time.Second)
			defer c()

			// Triger gracefull shutdown
			l.Info("gracefully shutting down http...")
			err := httpsrv.Shutdown(s)
			if err != nil {
				// Error from closing listeners, or context timeout:
				l.Error("error shutting down http", "err", err.Error())
			}
			l.Info("http shut down")
		}()

		err := httpsrv.ListenAndServe()
		if err != http.ErrServerClosed {
			// Error starting or closing listener:
			l.Error("failed to serve http", "err", err.Error())
		}
	}()

	// GRPC
	wg.Add(1)
	go func() {
		grpcl := l.With(slog.String("transport", "grpc"))
		grpcsrv := grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				grpcmw.RequestIDUnaryServerInterceptor, // before logger
				grpcmw.LoggerUnaryServerInterceptor(
					grpcmw.WithLogger(grpcl),
					grpcmw.WithConcise(true),
					grpcmw.WithLeak(false),
				// grpcmw.WithSensitive(map[string]struct{}{
				// 	"insecure":       {},
				// 	"very-insercure": {},
				// }),
				),
				grpcmw.RecoverUnaryServerInterceptor, // after error
			),
		)
		pb.RegisterHelloServiceServer(grpcsrv, hellogrpc.NewServer(grpcl, helloService))

		go func() {
			defer wg.Done()
			<-signalCtx.Done()

			l.Info("gracefully shutting down grpc...")
			grpcsrv.GracefulStop()
			l.Info("grpc shut down")
		}()

		lis, err := net.Listen("tcp", "0.0.0.0:8001")
		if err != nil {
			l.Error("grpc failed to listen", "err", err.Error())
		}

		if err := grpcsrv.Serve(lis); err != nil {
			l.Error("failed to serve grpc", "err", err.Error())
		}
	}()

	// PROMETHEUS
	wg.Add(1)
	go func() {
		r := mux.NewRouter()
		r.Path("/metrics").Handler(promhttp.Handler()).Methods("GET")

		promsrv := &http.Server{
			Handler: r,
			Addr:    "0.0.0.0:9000",
		}

		go func() {
			defer wg.Done()
			<-signalCtx.Done()

			s, c := context.WithTimeout(context.Background(), 30*time.Second)
			defer c()

			l.Info("gracefully shutting down promhttp...")
			err := promsrv.Shutdown(s)
			if err != nil {
				l.Error("error shutting down promhttp", "err", err.Error())
			}
			l.Info("promhttp shut down")
		}()

		err := promsrv.ListenAndServe()
		if err != http.ErrServerClosed {
			l.Error("failed to serve promhttp", "err", err.Error())
		}
	}()
	l.Info("started")

	wg.Wait()
}
