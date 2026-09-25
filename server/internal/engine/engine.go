package engine

import (
	"log/slog"

	"atlas.sageyutta.com/internal/gamedata"
)

type Engine struct {
	logger *slog.Logger
}

func NewEngine(logger *slog.Logger) *Engine {
	return &Engine{
		logger: logger,
	}
}

func (e *Engine) GetTreasureInfoFromId(id string) *gamedata.Treasure {
	treasures := gamedata.GetAllTreasures()
	for _, item := range treasures {
		if item.ID == id {
			return &item
		}
	}

	return nil
}
