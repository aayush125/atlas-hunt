package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"game.crisis-control.com/internal/gamedata"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Player struct {
	Nickname            string                   `json:"nickname" redis:"nickname"`
	GameID              string                   `json:"game_id" redis:"game_id"`
	JoinTimestamp       time.Time                `json:"join_timestamp" redis:"join_timestamp"`
	LastRejoinTimestamp time.Time                `json:"last_rejoin_timestamp" redis:"last_rejoin_timestamp"`
	PlayerID            uuid.UUID                `json:"id" redis:"id"`
	IsCreator           bool                     `json:"is_creator" redis:"is_creator"`
	SessionToken        string                   `json:"-" redis:"session_token"`
	TreasureSet         []gamedata.TreasureItem  `json:"treasure_set" redis:"treasure_set"`
	TreasureCount       map[string]int           `json:"treasure_count" redis:"treasure_count"`
	Inventory           []gamedata.InventoryItem `json:"inventory" redis:"inventory"`
	Points              int                      `json:"points" redis:"points"`
}

type PlayerModel struct {
	RDB *redis.Client
}

type PlayerInfo struct {
	Nickname string `json:"nickname"`
	ID       string `json:"id"`
}

func getPlayerKey(sessionToken string) string {
	return fmt.Sprintf("player:%s", sessionToken)
}

func (m *PlayerModel) BackupPlayerStats(token string, treasureSet []gamedata.TreasureItem, treasureCount map[string]int, inventory []gamedata.InventoryItem, points int) error {
	err := m.UpdateTreasureSet(token, treasureSet)
	if err != nil {
		return err
	}

	err = m.UpdateTreasureCount(token, treasureCount)
	if err != nil {
		return err
	}

	err = m.UpdateInventory(token, inventory)
	if err != nil {
		return err
	}

	err = m.UpdatePoints(token, points)
	if err != nil {
		return err
	}

	return nil
}

func (m *PlayerModel) Insert(player *Player) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := getPlayerKey(player.SessionToken)

	treasureData, err := json.Marshal(player.TreasureSet)
	if err != nil {
		return fmt.Errorf("marshalling treasure_set: %w", err)
	}

	treasureCountData, err := json.Marshal(player.TreasureCount)
	if err != nil {
		return fmt.Errorf("marshalling treasure_count: %w", err)
	}

	inventoryData, err := json.Marshal(player.Inventory)
	if err != nil {
		return err
	}

	if err := m.RDB.HSet(ctx, key, map[string]any{
		"nickname":              player.Nickname,
		"game_id":               player.GameID,
		"join_timestamp":        player.JoinTimestamp.Format(time.RFC3339),
		"last_rejoin_timestamp": player.LastRejoinTimestamp.Format(time.RFC3339),
		"id":                    player.PlayerID.String(),
		"is_creator":            player.IsCreator,
		"session_token":         player.SessionToken,
		"treasure_set":          treasureData,
		"treasure_count":        treasureCountData,
		"inventory":             inventoryData,
		"points":                player.Points,
	}).Err(); err != nil {
		return err
	}

	if err := m.RDB.Expire(ctx, key, 2*time.Hour).Err(); err != nil {
		return err
	}

	return nil
}

func (m *PlayerModel) SetTTL(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := getPlayerKey(token)

	if err := m.RDB.Expire(ctx, key, 20*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

func (m *PlayerModel) GetForSessionToken(token string) (*Player, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := getPlayerKey(token)

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

	var treasureSet []gamedata.TreasureItem
	if err := json.Unmarshal([]byte(data["treasure_set"]), &treasureSet); err != nil {
		return nil, fmt.Errorf("parsing treasure_set: %w", err)
	}

	var treasureCount map[string]int
	if err := json.Unmarshal([]byte(data["treasure_count"]), &treasureCount); err != nil {
		return nil, fmt.Errorf("parsing treasure_count: %w", err)
	}

	var inventory []gamedata.InventoryItem
	if err := json.Unmarshal([]byte(data["inventory"]), &inventory); err != nil {
		return nil, fmt.Errorf("parsing inventory: %w", err)
	}

	points, err := strconv.Atoi(data["points"])
	if err != nil {
		return nil, fmt.Errorf("parsing points: %w", err)
	}

	player := &Player{
		Nickname:            data["nickname"],
		GameID:              data["game_id"],
		JoinTimestamp:       joinTimestamp,
		LastRejoinTimestamp: lastRejoinTimestamp,
		PlayerID:            playerID,
		IsCreator:           isCreator,
		SessionToken:        data["session_token"],
		TreasureSet:         treasureSet,
		TreasureCount:       treasureCount,
		Inventory:           inventory,
		Points:              points,
	}

	return player, nil
}

func (m *PlayerModel) UpdateTreasureSet(token string, treasureSet []gamedata.TreasureItem) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	treasureData, err := json.Marshal(treasureSet)
	if err != nil {
		return fmt.Errorf("marshalling treasure_set: %w", err)
	}

	key := getPlayerKey(token)

	exists, err := m.RDB.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return ErrRecordNotFound
	}

	if err := m.RDB.HSet(ctx, key, "treasure_set", treasureData).Err(); err != nil {
		return fmt.Errorf("updating treasure_set: %w", err)
	}

	return nil
}

func (m *PlayerModel) UpdateInventory(token string, inventory []gamedata.InventoryItem) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	inventoryData, err := json.Marshal(inventory)
	if err != nil {
		return fmt.Errorf("marshalling inventory: %w", err)
	}

	key := getPlayerKey(token)

	exists, err := m.RDB.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return ErrRecordNotFound
	}

	if err := m.RDB.HSet(ctx, key, "inventory", inventoryData).Err(); err != nil {
		return fmt.Errorf("updating inventory: %w", err)
	}

	return nil
}

func (m *PlayerModel) UpdatePoints(token string, points int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := getPlayerKey(token)

	exists, err := m.RDB.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return ErrRecordNotFound
	}

	if err := m.RDB.HSet(ctx, key, "points", points).Err(); err != nil {
		return fmt.Errorf("updating points: %w", err)
	}

	return nil
}

func (m *PlayerModel) UpdateTreasureCount(token string, treasureCount map[string]int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := getPlayerKey(token)

	exists, err := m.RDB.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return ErrRecordNotFound
	}

	treasureCountData, err := json.Marshal(treasureCount)
	if err != nil {
		return fmt.Errorf("marshalling treasure_count: %w", err)
	}

	if err := m.RDB.HSet(ctx, key, "treasure_count", treasureCountData).Err(); err != nil {
		return fmt.Errorf("updating treasure_count: %w", err)
	}

	return nil
}

func (m *PlayerModel) Delete(sessionToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := getPlayerKey(sessionToken)

	if err := m.RDB.Del(ctx, key).Err(); err != nil {
		return err
	}

	return nil
}
