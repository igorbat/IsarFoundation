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

package scenario

import (
	"go-wml"
	"go-wesnoth/wesnoth"
)


type Scenario struct {
	id      string
	name    string
	body    string
	data    *wml.Data
	path    string
	defines []string
	Sides []string
}

func FromPath(id, path string, defines []string, prep wesnoth.Preprocessor) Scenario {
	body, err := prep.Preprocess(path, defines)
	if err != nil{
		panic ("Failed to prep "+err.Error())
	}
	mp := wml.ParseData (body)
	mpTags := mp.GetTags ("multiplayer")
	var data *wml.Data
	var found bool
	for _, m := range mpTags {
		if m.GetAttr("id") == id {
			found = true
			data = m
			
			break
		}
	}
	if !found {
		panic ("Couldn't find scenario "+ id)
	}
	name := data.GetAttr ("name")
	sides := data.GetTags ("side")
	playerSides := make ([]string, len(sides))
	for i, s := range sides {
		//assume sides are in order, but here is 0-indexed
		controller := s.GetAttr ("controller")
		s.AddAttr("side", i + 1)
		if controller == "" {
			controller = "none"
		}
		playerSides[i] = controller
	}
	bodyStr := data.String()
	return Scenario{id: id,body: bodyStr, name: name, path: path, data: data, defines: defines, Sides: playerSides}
}

func (s *Scenario) Id() string {
	return s.id
}

func (s *Scenario) Data() *wml.Data {
	return wml.ParseData ([]byte(s.body))//THIS IS BECAUSE THIS DATA ARE MANIPULATED, THEY SHOULDN'T CHANGE
}

func (s *Scenario) Name() string {
	return s.name
}

func (s *Scenario) Body() string {
	return s.body
}

func (s *Scenario) Path() string {
	return s.path
}

func (s *Scenario) Defines() []string {
	return s.defines
}
