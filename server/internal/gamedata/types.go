package gamedata

import "github.com/google/uuid"

const (
	CommonRarity    = "Common"
	RareRarity      = "Rare"
	EpicRarity      = "Epic"
	LegendaryRarity = "Legendary"
)

type Option map[string]string

type Question struct {
	ID      string `json:"id"`
	Q       string `json:"q"`
	Options Option `json:"options"`
}

type Treasure struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Icon        string     `json:"icon"`
	Cat         string     `json:"-"`
	Rarity      string     `json:"rarity"`
	Points      int        `json:"points"`
	BoostPoints int        `json:"boostPoints"`
	Questions   []Question `json:"questions"`
}

type InventoryItemBase struct {
	InstanceID  uuid.UUID `json:"instanceId"`
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	Cat         string    `json:"-"`
	Rarity      string    `json:"rarity"`
	Points      int       `json:"points"`
	BoostPoints int       `json:"boostPoints"`
}

type TreasureItem struct {
	InventoryItemBase
	X        float32  `json:"x"`
	Y        float32  `json:"y"`
	Question Question `json:"question"`
}

type InventoryItem struct {
	InventoryItemBase
}

// internal structs for parsing
type rawQuestion struct {
	ID      string `json:"id"`
	Q       string `json:"q"`
	Options Option `json:"options"`
	Answer  string `json:"answer"`
}

type rawTreasure struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Icon        string        `json:"icon"`
	Cat         string        `json:"cat"`
	Rarity      string        `json:"rarity"`
	Points      int           `json:"points"`
	BoostPoints int           `json:"boostPoints"`
	Questions   []rawQuestion `json:"questions"`
}

type Point struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type rawPoints struct {
	FirstMapPoints  []Point `json:"firstMapPoints"`
	SecondMapPoints []Point `json:"secondMapPoints"`
	ThirdMapPoints  []Point `json:"thirdMapPoints"`
}
