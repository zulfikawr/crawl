// cmd/crawl/main.go
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/zulfikawr/crawl/internal/cli"
)

func main() {
	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Warn().Msg("Interrupt received, shutting down gracefully...")
		os.Exit(0)
	}()

	// Execute CLI (app initialization happens inside cli.Execute)
	cli.Execute()
}
