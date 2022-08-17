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

package era

import (
	"fmt"
	"go-wesnoth/wesnoth"
	"go-wesnoth/resource"
	wml "go-wml"
	"strings"
)

type Era struct {
	Id       string
	Name     string
	Body     string
	Factions []*wml.Data
	Events   []*wml.Data
	Options  map[string]string
}

//var eras []byte

func Parse (id, path string, prep wesnoth.Preprocessor) Era {
	return ParseWithResources(id,path,prep,nil)
}

func ParseWithResources(id, path string, prep wesnoth.Preprocessor, res map[string]resource.Resource) Era {
	eras, err := prep.Preprocess(path, nil)
	if err != nil {
		panic ("Failed to prep "+err.Error())
	}
	fmt.Println("eras preprocess finished")
	
	erasData := wml.ParseData (eras)
	eraTags := erasData.GetTags ("era")
	var eraContent *wml.Data
	var found bool
	for _, e := range eraTags {
		if e.GetAttr("id") == id {
			found = true
			eraContent = e
			break
		}
	}
	if !found {
		panic ("Couldn't find era "+id)
	}

	name := eraContent.GetAttr ("name")
	
	factions := []*wml.Data{}
	factionTags := eraContent.GetTags ("multiplayer_side")
	for _, f := range factionTags {
		//delete (f, "description")
		//(eraContent["multiplayer_side"].(wml.Multiple))[i] = f
		if f.GetAttr("random_faction") != "yes" {
			factions = append (factions, f)
		}
	}
	newEraContent := wml.NewDataAttrs(eraContent.Attrs)
	for _, t := range eraContent.Tags {
		if t.Name != "load_resource" {
			newEraContent.Tags = append(newEraContent.Tags, t)
			continue
		}
		id := t.Data.GetAttr("id")
		r, ok := res[id]
		if !ok {
			panic("Can't find required resource "+id)
		}
		fmt.Println("Events", len(r.Events))
		for _, event := range r.Events {
			newEraContent.Tags = append(newEraContent.Tags, &wml.Tag {Name: "event", Data: event})
		}
	}
	eraContent = newEraContent
	body := wml.NewDataTags(&wml.Tag{Name: "era", Data: eraContent}).String()
	events := eraContent.GetTags ("event")
	options := map[string]string{}
	optTags := eraContent.GetTags("options")
	if len(optTags) == 1 {
		options = wesnoth.GetModOptions (optTags[0])
	}
	return Era{id, name, body, factions, events, options}
}

func LeaderPool (faction *wml.Data) []string {
	var leaders = []string{}
	if faction.Contains("random_leader") {
		leaders = strings.Split(faction.GetAttr("random_leader"), ",")
	} else {
		leaders = strings.Split(faction.GetAttr("leader"), ",")
	}
	return leaders
}
