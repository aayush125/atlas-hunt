package data

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

var (
	ErrRecordNotFound    = errors.New("record not found")
	ErrMaxPlayersReached = errors.New("max players reached")
)

type Models struct {
	Game   *GameModel
	Player *PlayerModel
}

func NewModels(rdb *redis.Client) *Models {
	return &Models{
		Game:   &GameModel{RDB: rdb},
		Player: &PlayerModel{RDB: rdb},
	}
}
