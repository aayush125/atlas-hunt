package gamedata

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

var ErrQuestionNotFound = errors.New("question not found")

func CheckAnswer(questionID, answer string) (bool, error) {
	correct, ok := answers[questionID]
	if !ok {
		return false, ErrQuestionNotFound
	}
	return answer == correct, nil
}

func getBonusPointsForPoints(points, count int, multiplier float32) int {
	factors := []float64{1, 1.5, 2.5, 3.2, 4}

	p := float64(points)
	m := float64(multiplier)

	round := func(x float64) int {
		return int(math.Ceil(math.Round(x*1e2) / 1e2))
	}

	if count <= len(factors) {
		if count == 1 {
			return points
		}
		return round(p * factors[count-1] * m)
	}

	overflow := float64(count - len(factors))
	return round((p * factors[len(factors)-1] * m) + (overflow * (p * 0.8 * m)))
}

func GetBonusPoints(treasureID string, treasureCount int) (int, error) {
	treasure, ok := treasureMap[treasureID]
	if !ok {
		return 0, fmt.Errorf("treasure not found")
	}

	switch strings.ToLower(treasure.Rarity) {
	case "common":
		return getBonusPointsForPoints(treasure.Points, treasureCount, 1), nil
	case "rare":
		return getBonusPointsForPoints(treasure.Points, treasureCount, 0.85), nil
	case "epic":
		return getBonusPointsForPoints(treasure.Points, treasureCount, 0.6), nil
	case "legendary":
		return getBonusPointsForPoints(treasure.Points, treasureCount, 0.55), nil
	default:
		return 0, fmt.Errorf("unknown treasure rarity")
	}
}

func TreasureExists(treasureId string) bool {
	_, ok := treasureMap[treasureId]
	return ok
}
