package engine

import (
	"crypto/rand"
	"math/big"

	"game.crisis-control.com/internal/gamedata"
	"github.com/google/uuid"
)

const (
	countCommon    = 20
	countRare      = 12
	countEpic      = 5
	countLegendary = 3
	countTotal     = countCommon + countRare + countEpic + countLegendary
)

func randIndex(n int) int {
	max := big.NewInt(int64(n))
	i, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	return int(i.Int64())
}

func (e *Engine) pickRandomQuestion(questions []gamedata.Question) gamedata.Question {
	return questions[randIndex(len(questions))]
}

func (e *Engine) pickRandomPoint(points []gamedata.Point) gamedata.Point {
	return points[randIndex(len(points))]
}

func (e *Engine) pickTreasures(pool []gamedata.Treasure, points []gamedata.Point, count int) []gamedata.TreasureItem {
	if count > len(pool) {
		panic("requested count exceeds treasure pool size")
	}

	items := make([]gamedata.TreasureItem, 0, count)

	for len(items) < count {
		t := pool[randIndex(len(pool))]
		point := e.pickRandomPoint(points)
		instanceID, err := uuid.NewRandom()
		if err != nil {
			e.logger.Error("error generating uuid for treasure instanceID", "error", err.Error())
			return nil
		}
		items = append(items, gamedata.TreasureItem{
			InventoryItemBase: gamedata.InventoryItemBase{
				InstanceID:  instanceID,
				ID:          t.ID,
				Name:        t.Name,
				Icon:        t.Icon,
				Cat:         t.Cat,
				Rarity:      t.Rarity,
				Points:      t.Points,
				BoostPoints: t.BoostPoints},
			X:        point.X,
			Y:        point.Y,
			Question: e.pickRandomQuestion(t.Questions),
		})
	}

	return items
}

func (e *Engine) GenerateTreasuresForPlayer() []gamedata.TreasureItem {
	items := make([]gamedata.TreasureItem, 0, countTotal)
	items = append(items, e.pickTreasures(gamedata.GetCommonTreasures(), gamedata.GetFirstMapPoints(), countCommon)...)
	items = append(items, e.pickTreasures(gamedata.GetRareTreasures(), gamedata.GetFirstMapPoints(), countRare)...)
	items = append(items, e.pickTreasures(gamedata.GetEpicTreasures(), gamedata.GetFirstMapPoints(), countEpic)...)
	items = append(items, e.pickTreasures(gamedata.GetLegendaryTreasures(), gamedata.GetFirstMapPoints(), countLegendary)...)
	return items
}

func (e *Engine) GenerateTreasureSet(players int) [][]gamedata.TreasureItem {
	sets := make([][]gamedata.TreasureItem, 0, players)
	for range players {
		sets = append(sets, e.GenerateTreasuresForPlayer())
	}
	return sets
}
