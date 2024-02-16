package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"

	"github.com/profclems/go-dotenv"
)

func main() {
	if err := dotenv.LoadConfig(); err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	builder := (&RouterBuilder{}).
		Port(dotenv.GetInt("PORT")).
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
