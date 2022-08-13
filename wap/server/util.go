// This file is part of Fastbot.
//
// Fastbot is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Fastbot is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Fastbot.  If not, see <https://www.gnu.org/licenses/>.

package server

import (
	"math/rand"
	//"regexp"
	"strings"
	"time"

	serverTypes "wap/server/types"
	//"go-wesnoth/wesnoth"
	"go-wml"
	"go-wesnoth/era"
)

var (
colors = []string{
		"red", "teal", "blue", "orange", "white","purple", "black","brown","green",
}
)

func SplitMessage(text string) []string {
	upperLimit := 256

	result := []string{}

	for pos := 0; pos < len(text); {
		from := pos
		var to int
		if pos+upperLimit < len(text) {
			to = pos + upperLimit
		} else {
			to = len(text)
		}
		result = append(result, text[from:to])
		pos = to
	}
	return result
}

func insertFaction(side *serverTypes.Side, faction *wml.Data, units map[string]*wml.Data) *wml.Data {
	var leaders = era.LeaderPool (faction)
	rand.Seed(time.Now().UTC().UnixNano())
	leader := leaders[rand.Int31n(int32(len(leaders)))]

	var gender string
	unit, ok := units[leader]
	if !ok {
		panic ("Leader's not found "+leader)
	}
	if (unit).Contains ("gender") {
		gender = unit.GetAttr("gender")
		if genders := strings.Split(gender, ","); len(genders) == 2 {
			rand.Seed(time.Now().UTC().UnixNano())
			gender = genders[rand.Int31n(2)]
		}
	} else {
		gender = "male"
	}
	insertTag := &wml.Tag{"insert", wml.NewDataAttrs(wml.AttrMap{
		"chose_random":   true,
		"color":          side.Color,
		"current_player": side.Player,
		"faction":        faction.GetAttr("id"),
		"faction_name":   faction.GetAttr("name"),
		"gender":         gender,
		"is_host":        false,
		"is_local":       false,
		"leader":         faction.GetAttr("leader"),
		"name":           side.Player,
		"player_id":      side.Player,
		"recruit":        faction.GetAttr("recruit"),
		"terrain_liked":  faction.GetAttr("terrain_liked"),
		"type":           leader,
		// Not necessary since already defined in the scenario:
		//"user_team_name": "whatever",
	})}
	if faction.Contains("random_leader") {
		insertTag.Data.AddAttr("random_leader", faction.GetAttr("random_leader"))
	}
	deleteTag := &wml.Tag{"delete", wml.NewDataAttrs(wml.AttrMap{"random_faction": "x"})}
	insert_childTag := &wml.Tag{"insert_child", &wml.Data{Attrs: wml.AttrMap{"index": 0}, Tags: []*wml.Tag{&wml.Tag{"ai",faction.GetTags("ai")[0]}}}} 
	sideData := wml.NewDataTags (insertTag,deleteTag,insert_childTag)
	return sideData
}

func getFactionID(faction *wml.Data) string {
	return faction.GetAttr("id")
}
