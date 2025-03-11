package tui

import (
	"github.com/PizzaHomicide/hisame/internal/config"
	"github.com/PizzaHomicide/hisame/internal/ui/tui/models"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(cfg *config.Config) error {
	p := tea.NewProgram(models.NewAppModel(cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
