package data

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Player struct {
	Nickname            string    `json:"nickname" redis:"nickname"`
	GameID              string    `json:"game_id" redis:"game_id"`
	JoinTimestamp       time.Time `json:"join_timestamp" redis:"join_timestamp"`
	LastRejoinTimestamp time.Time `json:"last_rejoin_timestamp" redis:"last_rejoin_timestamp"`
	PlayerID            uuid.UUID `json:"id" redis:"id"`
	IsCreator           bool      `json:"is_creator" redis:"is_creator"`
	SessionToken        string    `json:"-" redis:"session_token"`
}

type PlayerModel struct {
	RDB *redis.Client
}

func (m *PlayerModel) Insert(player *Player) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	key := fmt.Sprintf("player:%s", player.SessionToken)
	if err := m.RDB.HSet(ctx, key, map[string]any{
		"nickname":              player.Nickname,
		"game_id":               player.GameID,
		"join_timestamp":        player.JoinTimestamp,
		"last_rejoin_timestamp": player.LastRejoinTimestamp.Format(time.RFC3339),
		"id":                    player.PlayerID.String(),
		"is_creator":            player.IsCreator,
		"session_token":         player.SessionToken,
	}).Err(); err != nil {
		return err
	}
	if err := m.RDB.Expire(ctx, key, 30*time.Minute).Err(); err != nil {
		return err
	}
	return nil
}

func (m *PlayerModel) GetForSessionToken(token string) (*Player, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("player:%s", token)

	data, err := m.RDB.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, ErrRecordNotFound
	}

	joinTimestamp, err := time.Parse(time.RFC3339, data["join_timestamp"])
	if err != nil {
		return nil, fmt.Errorf("parsing join_timestamp: %w", err)
	}

	lastRejoinTimestamp, err := time.Parse(time.RFC3339, data["last_rejoin_timestamp"])
	if err != nil {
		return nil, fmt.Errorf("parsing last_rejoin_timestamp: %w", err)
	}

	playerID, err := uuid.Parse(data["id"])
	if err != nil {
		return nil, fmt.Errorf("parsing player id: %w", err)
	}

	isCreator, err := strconv.ParseBool(data["is_creator"])
	if err != nil {
		return nil, fmt.Errorf("parsing is_creator: %w", err)
	}

	player := &Player{
		Nickname:            data["nickname"],
		GameID:              data["game_id"],
		JoinTimestamp:       joinTimestamp,
		LastRejoinTimestamp: lastRejoinTimestamp,
		PlayerID:            playerID,
		IsCreator:           isCreator,
		SessionToken:        data["session_token"],
	}

	return player, nil
}
