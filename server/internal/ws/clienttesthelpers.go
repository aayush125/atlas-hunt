package ws

import (
	"fmt"
	"io"
	"log/slog"
	"testing"

	"atlas.sageyutta.com/internal/gamedata"
	"github.com/google/uuid"
)

const (
	FAKE_SESSION = "fake-session"
)

func initializeFakeClient(t testing.TB, room *Room, id string, isCreator bool) *Client {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return NewClient(
		nil,
		room,
		fmt.Sprintf("player-%s", id),
		id,
		FAKE_SESSION+"-"+id,
		logger,
		isCreator,
	)
}

func makeFakeInventory(t testing.TB, size int) []gamedata.InventoryItem {
	t.Helper()

	items := []gamedata.InventoryItem{}
	treasure := gamedata.GetAllTreasures()[0]
	for range size {
		items = append(items, gamedata.InventoryItem{
			InventoryItemBase: gamedata.InventoryItemBase{
				InstanceID:  uuid.New(),
				ID:          treasure.ID,
				Name:        treasure.Name,
				Icon:        treasure.Icon,
				Cat:         treasure.Cat,
				Rarity:      treasure.Rarity,
				Points:      treasure.Points,
				BoostPoints: treasure.BoostPoints,
			},
		})
	}

	return items
}
