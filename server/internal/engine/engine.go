package engine

import "log/slog"

type Engine struct {
	logger *slog.Logger
}

func NewEngine(logger *slog.Logger) *Engine {
	return &Engine{
		logger: logger,
	}
}
