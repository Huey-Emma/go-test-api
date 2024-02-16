package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type ErrResp struct {
	Detail string `json:"detail"`
}

type RespMsg struct {
	Message string `json:"message"`
}

type router struct {
	port   int
	logger *slog.Logger
}

type RouterBuilder struct {
	port   *int
	logger *slog.Logger
}

func (rb *RouterBuilder) Port(port int) *RouterBuilder {
	rb.port = &port
	return rb
}

func (rb *RouterBuilder) Logger(logger *slog.Logger) *RouterBuilder {
	rb.logger = logger
	return rb
}

func (rb *RouterBuilder) Build() (*router, error) {
	if rb.logger == nil {
		rb.logger = slog.Default()
	}

	port := 8000

	if rb.port == nil || *rb.port == 0 {
		rb.port = &port
	}

	return &router{
		port:   *rb.port,
		logger: rb.logger,
	}, nil
}

func (router *router) handler() http.Handler {
	mux := http.DefaultServeMux

	mux.HandleFunc("GET /", router.adapter(router.Home))

	return router.Recover(mux)
}

func (router router) JSON(w http.ResponseWriter, code int, v any, header http.Header) error {
	w.Header().Set("Content-Type", "application/json")
	for k, v := range header {
		w.Header()[k] = v
	}

	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

type fn func(http.ResponseWriter, *http.Request) error

func (router) adapter(fn fn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			panic(err)
		}
	}
}

func (router *router) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				router.logger.Error("server error", slog.String("detail", fmt.Sprintf("%v", r)))
				_ = router.JSON(
					w,
					http.StatusInternalServerError,
					ErrResp{Detail: "server error"},
					nil,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (router *router) Home(w http.ResponseWriter, r *http.Request) error {
	return router.JSON(
		w, http.StatusOK,
		RespMsg{Message: "Dockerizing a Go application"},
		nil,
	)
}

func (router *router) Listen(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", router.port),
		Handler: router.handler(),
	}

	errch := make(chan error, 1)

	go func() {
		router.logger.Info("server is starting", slog.Int("port", router.port))
		if err := server.ListenAndServe(); err != nil {
			errch <- err
		}
	}()

	select {
	case err := <-errch:
		return err
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		router.logger.Info("server is shutting down")
		err := server.Shutdown(ctx)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}
