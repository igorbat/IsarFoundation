package main

import (
	b "newwesbot"
	"strings"
	"math"
)

type AfterParams struct {
}

var (
	_ b.LadderParameters = AfterParams{}
)

func (_ AfterParams) HasWon (pls []string, team int, pl string) bool {
	if team == 0 {
		return false
	}
	for i, p := range pls {
		if p == strings.ToLower (pl) {
			return team - 1 == i
		}
	}
	return false
}

func (_ AfterParams) HasLost (pls []string, team int, pl string) bool {
	if team == 0 {
		return false
	}
	for i, p := range pls {
		if p == strings.ToLower (pl) {
			return team - 1 != i
		}
	}
	return false
}

func (_ AfterParams) TeamFun (pls []string, reporter string) (int, int) {
	id := -1
	for i, p := range pls {
		if p == strings.ToLower (reporter) {
			id = i
			break
		}
	}
	if id == -1 {
		return -1, -1
	}
	return id + 1, 2 - id
}

func (_ AfterParams) RatingFun (rats []int, teamWon int) []int {
	winner := teamWon - 1
	points := []float64{0.0, 0.0}
	points[winner] = 1.0
	expected := []float64 {0.0, 0.0}
	expected[0] = 1.0 / (1.0 + math.Pow (10.0, float64(rats[1] - rats[0])/400.0))
	expected[1] = 1.0 / (1.0 + math.Pow (10.0, float64(rats[0] - rats[1])/400.0))
	adjs := []int {0, 0}
	for i, _ := range expected {
		k := 30.0
		if rats[i] > 1800 {
			k = 10.0
		} else if rats[i] > 1650{
			k = 15.0
		} else if rats[i] > 1400 {
			k = 20.0
		}
		adjs[i] = int (k * (points[i] - expected[i]))
		if adjs[i] == 0 {
			if i == winner {
				adjs[i] = 1//at least one point
			} else {
				adjs[i] = -1
			}
		}
	}
	return adjs
}
