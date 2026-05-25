package gamedata

import (
	_ "embed"
	"encoding/json"
	"log"
)

//go:embed treasures.json
var treasureData []byte

//go:embed points.json
var pointsData []byte

var treasures []Treasure
var points rawPoints

var commonTreasures []Treasure
var rareTreasures []Treasure
var epicTreasures []Treasure
var legendaryTreasures []Treasure

var answers map[string]string       // question id -> correct answer
var treasureMap map[string]Treasure // treasure id -> treasure

func init() {
	var raw []rawTreasure
	if err := json.Unmarshal(treasureData, &raw); err != nil {
		log.Fatalf("failed to parse treasures.json: %v", err)
	}

	if err := json.Unmarshal(pointsData, &points); err != nil {
		log.Fatalf("failed to parse points.json: %v", err)
	}

	answers = make(map[string]string)
	treasureMap = make(map[string]Treasure)

	for _, r := range raw {
		questions := make([]Question, len(r.Questions))
		for i, q := range r.Questions {
			answers[q.ID] = q.Answer
			questions[i] = Question{
				ID:      q.ID,
				Q:       q.Q,
				Options: q.Options,
			}
		}
		treasure := Treasure{
			ID:          r.ID,
			Name:        r.Name,
			Icon:        r.Icon,
			Cat:         r.Cat,
			Rarity:      r.Rarity,
			Points:      r.Points,
			BoostPoints: r.BoostPoints,
			Questions:   questions,
		}

		treasures = append(treasures, treasure)
		treasureMap[treasure.ID] = treasure

		switch r.Rarity {
		case CommonRarity:
			commonTreasures = append(commonTreasures, treasure)
		case RareRarity:
			rareTreasures = append(rareTreasures, treasure)
		case EpicRarity:
			epicTreasures = append(epicTreasures, treasure)
		case LegendaryRarity:
			legendaryTreasures = append(legendaryTreasures, treasure)
		default:
			log.Fatalf("unknown rarity %q for treasure %s in treasures.json", r.Rarity, r.ID)
		}
	}
}
