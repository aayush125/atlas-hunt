package gamedata

func GetAllTreasures() []Treasure {
	result := make([]Treasure, len(treasures))
	copy(result, treasures)
	return result
}

func GetCommonTreasures() []Treasure {
	result := make([]Treasure, len(commonTreasures))
	copy(result, commonTreasures)
	return result
}

func GetRareTreasures() []Treasure {
	result := make([]Treasure, len(rareTreasures))
	copy(result, rareTreasures)
	return result
}

func GetEpicTreasures() []Treasure {
	result := make([]Treasure, len(epicTreasures))
	copy(result, epicTreasures)
	return result
}

func GetLegendaryTreasures() []Treasure {
	result := make([]Treasure, len(legendaryTreasures))
	copy(result, legendaryTreasures)
	return result
}

func GetFirstMapPoints() []Point {
	result := make([]Point, len(points.FirstMapPoints))
	copy(result, points.FirstMapPoints)
	return result
}

func GetSecondMapPoints() []Point {
	result := make([]Point, len(points.SecondMapPoints))
	copy(result, points.SecondMapPoints)
	return result
}

func GetThirdMapPoints() []Point {
	result := make([]Point, len(points.ThirdMapPoints))
	copy(result, points.ThirdMapPoints)
	return result
}
