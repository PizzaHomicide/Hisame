package main

import (
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/config"
	"github.com/PizzaHomicide/hisame/internal/log"
	"github.com/PizzaHomicide/hisame/internal/ui/tui"
	"github.com/PizzaHomicide/hisame/internal/version"
	"os"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		// It is unrecoverable if we cannot produce an application config
		_, _ = fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialise logger
	logger, err := log.New(log.Config{
		Level:    cfg.Logging.Level,
		FilePath: cfg.Logging.FilePath,
	})
	if err != nil {
		// Probably should let the app continue without logging, but for now this is acceptable.
		_, _ = fmt.Fprintf(os.Stderr, "failed to initialise logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	// Set the default global logger
	log.SetDefaultLogger(logger)

	log.Info("Starting up Hisame", "version", version.GetVersion(), "build_time", version.GetBuildTime())

	if err := tui.Run(cfg); err != nil {
		log.Error("Unhandled error while running TUI", "error", err)
		os.Exit(1)
	}

	log.Info("Hisame shutting down.  Goodbye!")
}
