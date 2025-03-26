package player

import (
	"fmt"

	"github.com/PizzaHomicide/hisame/internal/config"
	"github.com/PizzaHomicide/hisame/internal/log"
)

// CreateVideoPlayer creates a new video player based on the configuration
func CreateVideoPlayer(cfg *config.Config) (VideoPlayer, error) {
	playerType := cfg.Player.Type
	log.Info("Creating video player", "type", playerType)

	switch playerType {
	case "mpv":
		return NewMPVPlayer(cfg), nil
	case "custom":
		// Custom player implementation (future extension)
		return nil, fmt.Errorf("custom player not yet implemented")
	default:
		log.Warn("Unknown player type, falling back to MPV", "type", playerType)
		return NewMPVPlayer(cfg), nil
	}
}
