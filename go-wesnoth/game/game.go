// This file is part of Go Wesnoth.
//
// Go Wesnoth is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Go Wesnoth is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Go Wesnoth.  If not, see <https://www.gnu.org/licenses/>.

package game

import (
	"go-wesnoth/era"
	"go-wesnoth/mod"
	"go-wesnoth/addon"
	"go-wesnoth/scenario"
	"go-wml"
	"strings"
	"strconv"
	//"fmt"
)

var (
	sideData = &wml.Data{
		Attrs: wml.AttrMap {
			"allow_changes":   true,
			"allow_player":    true,
			"chose_random":    false,
			"faction":         "Random",
			"faction_name":    wml.Domain{wml.Tr("Random"), "wesnoth-multiplayer"},
			"fog":             true,
			"gender":          "null",
			"gold":            100,
			"income":          0,
			"is_host":         false,
			"is_local":        false,
			"random_faction":  true,
			"shroud":          false,
			"type":            "null",
			"village_gold":    2,
			"village_support": 1,
		},
		Tags: []*wml.Tag{wml.EmptyTag("default_faction")},
		// Must be defined inside a real scenario:
		//"canrecruit":     true,
		//"controller":     "human",
		//"side":            1,
		//"team_name":       "whatever",
		//"user_team_name":  "whatever",
		//"ai":              wml.Data{"villages_per_scout": 8},
		// Must be manually defined:
		//"color":           "red",
	}
	
	colors = []string{
		"red", "blue", "green", "orange", "teal","purple", "black","brown","white",
	}
)

type Game struct {
	Title      string
	Scenario   scenario.Scenario
	Era        era.Era
	Mods       []mod.Mod
	Addons     []addon.Addon
	Version    string
	NotNewGame bool   // To set up manually
	Id         string // Obtained by Parse()
	Name       string // Obtained by Parse()
	// Timer-related config
	TimerEnabled  bool
	InitTime      int
	TurnBonus     int
	ReservoirTime int
	ActionBonus   int
	ExtraVariables map[string]string
	expModifier   int
}

func NewGame(title string, scenario scenario.Scenario,
	era era.Era, mods []mod.Mod, addons []addon.Addon,
	timerEnabled bool, initTime int, turnBonus int, reservoirTime int,
	actionBonus int, version string) Game {
	game := Game{Title: title, Scenario: scenario,
		Era: era,
		Mods: mods,
		Addons: addons,
		Version: version,
		TimerEnabled: timerEnabled, InitTime: initTime, TurnBonus: turnBonus,
		ReservoirTime: reservoirTime, ActionBonus: actionBonus}
	game.Parse()
	return game
}

func (g *Game) Parse() {
	g.Id = g.Scenario.Id()
	g.Name = g.Scenario.Name()
}

func (g *Game) Bytes() []byte {
	return []byte(g.String())
}

func (g *Game) String() string {
	return g.topLevel() +
		g.scenarioBlock() +
		g.carryoverBlock() +
		g.multiplayerBlock() + g.eraBlock() + g.modsBlock()
}

func (g *Game) topLevel() string {
	topLevel := &wml.Data{
		Attrs: wml.AttrMap {
			"abbrev":                 "",
			"campaign":               "",
			"campaign_define":        "",
			"campaign_extra_defines": "",
			"campaign_name":          "",
			"campaign_type":          "multiplayer",
			"difficulty":             "NORMAL",
			"end_credits":            true,
			"end_text":               "",
			"end_text_duration":      0,
			"era_define":             "",
			"era_id":                 g.Era.Id,
			"label":                  g.Name,
			"mod_defines":            "",
			"oos_debug":              false,
			"random_mode":            "",
			"scenario_define":        "",
			"version":                g.Version,
		},
		Tags: []*wml.Tag{&wml.Tag{Name:"replay", Data: wml.NewDataTags(wml.EmptyTag ("upload_log"))}},
	}
	return topLevel.String()
}

func (g *Game) scenarioBlock() string {
	data := g.Scenario.Data()
	col := -1
	sideTags := data.GetTags ("side")
	for _, sVal := range sideTags {
		col++
		if sVal.GetAttr("controller") != "human" {
			continue
		}
		sideAttrs, sideTags := sideTag(wml.MergeData(sideData, sVal), colors[col%len(colors)], "")
		sVal.Attrs = sideAttrs
		sVal.Tags = sideTags
		
	}
	//fmt.Println ("ELNINININININIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII", data)

	if g.NotNewGame == true {
		data.AddAttr("allow_new_game", false)
		data.AddAttr("disallow_recall", true)
	}
	var err error
	g.expModifier, err = strconv.Atoi (data.GetAttr("experience_modifier"))
	if err != nil {
		g.expModifier = 70
		data.AddAttr("experience_modifier", 70)
	}
	data.AddAttr("has_mod_events", true)
	if data.GetAttr("objectives") == "" {
		data.AddAttr("objectives", "<big>Victory:</big>\n<span color='#00ff00'>• Defeat enemy leader(s)</span>")
	}
	data.AddAttr("turns", -1)
	for _, v := range g.Era.Events {
		data.AddTagByName ("event", v)
	}
	for _, m := range g.Mods {
		for _, v := range m.Events {
			data.AddTagByName ("event", v)
		}
		for _, v := range m.ModUnitTypes {
			data.AddTagByName ("modify_unit_type", v)
		}
	}
	
	finalData := wml.NewData()//wml.Data {"scenario": data}
	finalData.AddTagByName ("scenario", data)
	//fmt.Println ("#############################################################", finalData, "################################3\n", data)

	return finalData.String()
}

func (g *Game) carryoverBlock() string {
	variables := wml.NewData()
	for _, mod := range g.Mods {
		if len(mod.Options) > 0 {
			for optname, optval := range mod.Options {
				variables.AddAttr(optname, optval)
			}
			
		}
	}
	if len(g.ExtraVariables) > 0 {
		for extraName, extraVal := range g.ExtraVariables {
			variables.AddAttr (extraName, extraVal)
		}
	}
	carryover := &wml.Tag{"carryover_sides_start", &wml.Data{
		Attrs: wml.AttrMap{
			"next_scenario": g.Id,
			"random_calls":  0,
			"random_seed":   randomSeed(),
		},
		Tags: []*wml.Tag{&wml.Tag{"variables", variables}} ,
	}}
	return carryover.String()
}

func (g *Game) multiplayerBlock() string {
	options := wml.NewData()
	if len(g.Era.Options) > 0 {
		eraOpts := wml.NewData()
		for optname, optval := range g.Era.Options {
			eraOpts.AddTagByName ("option", wml.NewDataAttrs(wml.AttrMap{"id": optname, "value": optval}))
		}
		eraOpts.AddAttr("id", g.Era.Id)
		options.AddTagByName("era", eraOpts)
	}
	modIds := []string{}
	for _, mod := range g.Mods {
		modIds = append(modIds, mod.Id)
		if len(mod.Options) > 0 {
			modOpts := wml.NewData()
			for optname, optval := range mod.Options {
				modOpts.AddTagByName ("option", wml.NewDataAttrs(wml.AttrMap{"id": optname, "value": optval}))
			}
			modOpts.AddAttr("id", mod.Id)
			options.AddTagByName ("modification", modOpts)
		}
	}
	mpContents := &wml.Data{
		Attrs: wml.AttrMap {
			"active_mods":                 strings.Join (modIds, ","),
			"difficulty_define":           "NORMAL",
			"experience_modifier":         g.expModifier,
			"hash":                        "",
			"mp_campaign":                 "",
			"mp_campaign_name":            "",
			"mp_countdown":                g.TimerEnabled,
			"mp_countdown_action_bonus":   g.ActionBonus,
			"mp_countdown_init_time":      g.InitTime,
			"mp_countdown_reservoir_time": g.ReservoirTime,
			"mp_countdown_turn_bonus":     g.TurnBonus,
			"mp_era":                      g.Era.Id,
			"mp_era_name":                 g.Era.Name,
			"mp_fog":                      true,
			"mp_num_turns":                -1,
			"mp_random_start_time":        false,
			"mp_scenario":                 g.Id,
			"mp_scenario_name":            g.Name,
			"mp_shroud":                   false,
			"mp_use_map_settings":         true,
			"mp_village_gold":             2,
			"mp_village_support":          1,
			"observer":                    true,
			"random_faction_mode":         "No Mirror",
			"registered_users_only":       false,
			"savegame":                    false,
			"scenario":                    g.Title,
			"shuffle_sides":               true,//false
			"side_users":                  "",
		},
		Tags: []*wml.Tag{&wml.Tag{"options",options}},
	}
	for _, addon := range g.Addons {
		mpContents.AddTagByName ("addon", addon.ToWML())
	}
	multiplayer := wml.Tag{"multiplayer", mpContents}
	return multiplayer.String()
}

func (g *Game) eraBlock() string {
	return g.Era.Body
}

func (g *Game) modsBlock() string {
	modsBodies := make([]string, len (g.Mods))
	for i, v := range g.Mods {
		modsBodies[i] = v.Body
	}
	return strings.Join (modsBodies, "\n")
}
