package main

import (
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/config"
	"github.com/PizzaHomicide/hisame/internal/log"
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

	log.Info("Hisame starting", "version", "0.0.1")

	log.Debug("This is a sample hisame debug log", "test", true)
	log.Info("This is a sample hisame info log", "test", true)
	log.Warn("This is a sample hisame warning log", "test", true)
	log.Error("This is a sample hisame error log", "test", true)

	log.Info("Hisame shutting down.  Goodbye!")
}
