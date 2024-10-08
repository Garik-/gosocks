package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Garik-/gosocks/internal/app"
	_ "go.uber.org/automaxprocs"
)

func slogLevel(verbose bool) slog.Level {
	if verbose {
		return slog.LevelDebug
	}

	return slog.LevelInfo
}

func main() {
	var verbose bool

	const (
		verboseUsage   = "show verbose debug information"
		verboseDefault = true
	)

	flag.BoolVar(&verbose, "v", verboseDefault, verboseUsage)
	flag.BoolVar(&verbose, "verbose", verboseDefault, verboseUsage)

	address := flag.String("address", ":8080", "srv address")

	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel(verbose),
	})))

	slog.Info("init",
		slog.String("address", *address),
		slog.Bool("verbose", verbose),
	)

	srv, err := app.NewServer(*address)
	if err != nil {
		slog.Error("newServer", slog.String("err", err.Error()))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	srv.Start(ctx)

	// Wait for a SIGINT or SIGTERM signal to gracefully shut down the srv
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	cancel()
	slog.Debug("shutting down srv...")
	srv.Stop(time.Second)
	slog.Info("srv stopped")
}
