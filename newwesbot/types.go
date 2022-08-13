package newwesbot

import (
	"strings"
)

type Player struct {
	Name string `pg:",pk"`
	Rating int
	Banned bool `pg:",notnull,use_zero"`
}

type Game struct {
	Id int 
	StartTime int64
	Players []string `pg:",array"`
	Ratings []int `pg:",array"`
	TeamWon int `pg:",notnull,use_zero"`//0 is played
	Reporter string
	Contested bool `pg:",notnull,use_zero"`
	Saved bool `pg:",notnull,use_zero"`
	ContestedBy string
	Canceled bool `pg:",notnull,use_zero"`
	RatAdjusts []int
}

func (game Game) HasPlayed (pl string) bool {
	for _, p := range game.Players {
		if p == strings.ToLower (pl) {
			return true
		}
	}
	return false
}

type LadderParameters interface {
	RatingFun ([]int, int) []int
	TeamFun ([]string, string) (int, int)
	HasWon ([]string, int, string) bool
	HasLost ([]string, int, string) bool
}
