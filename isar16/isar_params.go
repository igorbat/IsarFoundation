package main

import (
	b "newwesbot"
	"strings"
)

type IsarParams struct {
}

var (
	_ b.LadderParameters = IsarParams{}
)

func (_ IsarParams) HasWon (pls []string, team int, pl string) bool {
	if team == 0 {
		return false
	}
	for i, p := range pls {
		if p == strings.ToLower (pl) {
			if team == 1 {
				return i == 0 || i == 3
			} else {
				return i == 1 || i == 2
			}
		}
	}
	return false
}

func (_ IsarParams) HasLost (pls []string, team int, pl string) bool {
	if team == 0 {
		return false
	}
	for i, p := range pls {
		if p == strings.ToLower (pl) {
			if team == 2 {
				return i == 0 || i == 3
			} else {
				return i == 1 || i == 2
			}
		}
	}
	return false
}

func (_ IsarParams) TeamFun (pls []string, reporter string) (int, int) {
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
	if id == 0 || id == 3 {
		return 1,2
	} else {
		return 2,1
	}
}

func (_ IsarParams) RatingFun (rats []int, teamWon int) []int {
	if len (rats) != 4 {
		panic("Rating fail")
	}
	var winners, losers []int
	if teamWon == 1 {
		winners = []int{0, 3}
		losers = []int{1, 2}
	} else {
		winners = []int{1, 2}
		losers = []int{0, 3}
	}
	winRate := rats[winners[0]] + rats[winners[1]]
	loseRate := rats[losers[0]] + rats[losers[1]]
	diff := winRate - loseRate
	pie := 10
	switch {
		case diff <= -300:
			pie = 30
		case diff <= -100:
			pie = 25
		case diff <= 100:
			pie = 20
		case diff <= 300:
			pie = 15
	}
	adjs := []int {0, 0, 0, 0}
	adjs[winners[0]] = (pie * rats[winners[0]]) / winRate
	adjs[winners[1]] = (pie * rats[winners[1]]) / winRate
	adjs[losers[0]] = -(pie * rats[losers[0]]) / loseRate
	adjs[losers[1]] = -(pie * rats[losers[1]]) / loseRate
	return adjs
}
