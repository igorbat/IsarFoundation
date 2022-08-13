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

package mod

import (
	"fmt"
	"go-wesnoth/wesnoth"
	"go-wml"
)

type Mod struct {
	Id       string
	Name     string
	Body     string
	Events   []*wml.Data
	ModUnitTypes []*wml.Data
	Options  map[string]string
}


func Parse(id, path string, prep wesnoth.Preprocessor) Mod {
	mods, err := prep.Preprocess(path, nil)
	if err != nil {
		panic ("Failed to prep "+err.Error())
	}
	fmt.Println("mods preprocess finished")
	
	modsData := wml.ParseData (mods)
	modsTags := modsData.GetTags ("modification")
	var modContent *wml.Data
	var found bool
	for _, m := range modsTags {
		if m.GetAttr("id") == id {
			found = true
			modContent = m
			break
		}
	}
	if !found {
		panic ("Couldn't find mod "+id)
	}

	name := modContent.GetAttr ("name")
	
	
	body := wml.NewDataTags(&wml.Tag{Name: "modification", Data: modContent}).String()
	events := modContent.GetTags ("event")
	modunittypes := modContent.GetTags ("modify_unit_type")
	options := map[string]string{}
	optTags := modContent.GetTags("options")
	if len(optTags) == 1 {
		options = wesnoth.GetModOptions (optTags[0])
	}
	return Mod{id, name, body, events, modunittypes, options}
}
