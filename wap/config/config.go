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

package config

import (
	"io/ioutil"
	"encoding/json"
	"go-wesnoth/addon"
)


type Config struct {
	Hostname string
	Port uint16
	Timer TimerConfig
	Admins []string
	Username string
	Password string
	DBName string
	DBUser string
	DBPass string
	EraId string
	EraPath string
	ScenarioId string
	ScenarioPath string
	ModPaths map[string]string
	ResPaths []string
	UnitsPath string
	WesnothBinary string
	WesnothData string
	GameTitle string
	GreetMessage string
	ExtraMessage string
	UnqualifiedMessage string
	NoTLS bool
	Addons []addon.Addon
}

func LoadConfig (path string) (conf Config) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic (err)
	}
	
	err = json.Unmarshal (data, &conf)
	if err != nil {
		panic (err)
	}
	return
}
type GameConfig struct {
	Title         string
	Players       []string
	PickingPlayer string
	Scenarios     []ScenarioConfig
}

type ScenarioConfig struct {
	Path    string
	Defines []string
}

type TimerConfig struct {
	Enabled       bool
	InitTime      int
	TurnBonus     int
	ReservoirTime int
	ActionBonus   int
}
