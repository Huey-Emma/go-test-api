package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	builder := (&RouterBuilder{}).
		Port(8000).
		Logger(logger)

	router, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}

	sigs := []os.Signal{os.Kill, os.Interrupt}
	ctx, cancel := signal.NotifyContext(context.Background(), sigs...)
	defer cancel()

	if err := router.Listen(ctx); err != nil {
		log.Fatal(err)
	}
}
