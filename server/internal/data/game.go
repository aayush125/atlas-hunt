package data

import (
	"context"
	"fmt"
	"time"

	"game.crisis-control.com/internal/validator"
	"github.com/redis/go-redis/v9"
)

type GameState string

func (g GameState) MarshalBinary() ([]byte, error) {
	return []byte(g), nil
}

func (g *GameState) UnmarshalBinary(data []byte) error {
	*g = GameState(data)
	return nil
}

const (
	GameStateLobby      GameState = "lobby"
	GameStateInProgress GameState = "in_progress"
)

type Game struct {
	MaxPlayerCount     int       `json:"max_players" redis:"max_players"`
	CurrentPlayerCount int       `json:"current_player_count" redis:"current_player_count"`
	GameID             string    `json:"id" redis:"id"`
	CreatedAt          time.Time `json:"created_at" redis:"created_at"`
	CreatorNickname    string    `json:"creator_nickname" redis:"creator_nickname"`
	State              GameState `json:"state" redis:"state"`
}

type JoinGame struct {
	Nickname string `json:"nickname"`
	GameCode string `json:"game_code"`
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

func validateGameCode(v *validator.Validator, code string) {
	v.Check(code != "", "game_code", "Game code is required to join a game.")
	v.Check(len(code) == 6, "game_code", "Game code must be exactly 6 characters long.")
}

func ValidateCreateGame(v *validator.Validator, game *Game) {
	validateNickname(v, game.CreatorNickname)
	validateNumberOfPlayers(v, game.MaxPlayerCount)
}

func ValidateJoinGame(v *validator.Validator, request *JoinGame) {
	validateNickname(v, request.Nickname)
	validateGameCode(v, request.GameCode)
}

type GameModel struct {
	RDB *redis.Client
}

func getGameKey(code string) string {
	return fmt.Sprintf("game:%s", code)
}

func getGamePlayerKey(code string) string {
	return fmt.Sprintf("game:%s:players", code)
}

func (m *GameModel) Insert(game *Game) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := getGameKey(game.GameID)

	if err := m.RDB.HSet(ctx, key, game).Err(); err != nil {
		return err
	}

	if err := m.RDB.Expire(ctx, key, 30*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

func (m *GameModel) GetForCode(gameCode string) (*Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := getGameKey(gameCode)

	cmd := m.RDB.HGetAll(ctx, key)
	if err := cmd.Err(); err != nil {
		return nil, err
	}

	if len(cmd.Val()) == 0 {
		return nil, ErrRecordNotFound
	}

	var game Game
	if err := cmd.Scan(&game); err != nil {
		return nil, err
	}

	return &game, nil
}

func (m *GameModel) Delete(gameCode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := getGameKey(gameCode)

	if err := m.RDB.Del(ctx, key).Err(); err != nil {
		return err
	}

	return nil
}

func (m *GameModel) IncrementPlayerCount(gameCode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	game, err := m.GetForCode(gameCode)
	if err != nil {
		return err
	}

	if game.CurrentPlayerCount >= game.MaxPlayerCount {
		return ErrMaxPlayersReached
	}

	key := getGameKey(gameCode)

	return m.RDB.HSet(ctx, key, "current_player_count", game.CurrentPlayerCount+1).Err()
}

func (m *GameModel) DecrementPlayerCount(gameCode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	game, err := m.GetForCode(gameCode)
	if err != nil {
		return err
	}

	if game.CurrentPlayerCount <= 0 {
		return ErrNoPlayersInGame
	}

	key := getGameKey(gameCode)

	return m.RDB.HSet(ctx, key, "current_player_count", game.CurrentPlayerCount-1).Err()
}

func (m *GameModel) AddPlayerToSet(gameCode, playerSessionToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := getGamePlayerKey(gameCode)
	if err := m.RDB.SAdd(ctx, key, playerSessionToken).Err(); err != nil {
		return err
	}

	if err := m.RDB.Expire(ctx, key, 30*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

func (m *GameModel) RemovePlayerFromSet(gameCode, playerSessionToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := getGamePlayerKey(gameCode)

	return m.RDB.SRem(ctx, key, playerSessionToken).Err()
}

func (m *GameModel) GetAllPlayers(gameCode string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := getGamePlayerKey(gameCode)

	players, err := m.RDB.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(players) == 0 {
		return nil, ErrRecordNotFound
	}

	return players, nil
}
