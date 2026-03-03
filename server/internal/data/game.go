package data

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"game.crisis-control.com/internal/validator"
	"github.com/redis/go-redis/v9"
)

type Game struct {
	NumberOfPlayers    int       `json:"number_of_players"`
	CurrentPlayerCount int       `json:"current_player_count"`
	GameID             string    `json:"id"`
	CreatedAt          time.Time `json:"created_at"`
	CreatorNickname    string    `json:"creator_nickname"`
}

func validateNickname(v *validator.Validator, nickname string) {
	v.Check(nickname != "", "nickname", "A nickname is required")
	v.Check(len(nickname) <= 12, "nickname", "Nickname must be 12 letters or less")
}

func validateNumberOfPlayers(v *validator.Validator, numberOfPlayers int) {
	v.Check(numberOfPlayers > 0, "number_of_players", "Number of players must be non-zero")
	v.Check(numberOfPlayers%2 == 0, "number_of_players", "Number of players must be even")
	v.Check(numberOfPlayers >= 6 && numberOfPlayers <= 10, "number_of_players", "Number of players must be between 6 and 10")
}

func ValidateCreateGame(v *validator.Validator, game *Game) {
	validateNickname(v, game.CreatorNickname)
	validateNumberOfPlayers(v, game.NumberOfPlayers)
}

type GameModel struct {
	RDB *redis.Client
}

func (m *GameModel) Insert(game *Game) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("game:%s", game.GameID)

	if err := m.RDB.HSet(ctx, key, map[string]any{
		"id":                   game.GameID,
		"max_players":          game.NumberOfPlayers,
		"current_player_count": game.CurrentPlayerCount,
		"created_at":           game.CreatedAt.Format(time.RFC3339),
		"creator_nickname":     game.CreatorNickname,
	}).Err(); err != nil {
		return err
	}

	if err := m.RDB.Expire(ctx, key, 30*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

func (m *GameModel) AddPlayerToGame(gameID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("game:%s", gameID)
	game, err := m.RDB.HGetAll(ctx, key).Result()
	if err != nil {
		return err
	}

	if len(game) == 0 {
		return ErrRecordNotFound
	}

	currentPlayerCount, err := strconv.Atoi(game["current_player_count"])
	if err != nil {
		return err
	}

	maxPlayers, err := strconv.Atoi(game["max_players"])
	if err != nil {
		return err
	}

	if currentPlayerCount >= maxPlayers {
		return ErrMaxPlayersReached
	}

	err = m.RDB.HSet(ctx, key, "current_player_count", currentPlayerCount+1).Err()
	if err != nil {
		return err
	}

	return nil
}
