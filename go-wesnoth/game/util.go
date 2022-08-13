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
	"math/rand"
	"time"
	"fmt"

	"go-wml"
)

func sideTag(data *wml.Data, color string, player string) (wml.AttrMap, []*wml.Tag) {
	side := wml.NewData()
	side.AddAttr("color", color)
	if player != "" {
		side.AddAttr("current_player", player)
		side.AddAttr("name", player)
		side.AddAttr("player_id", player)
	}
	side = wml.MergeData(data, side)
	return side.Attrs, side.Tags
}

func randomSeed() string {
	rand.Seed(time.Now().UTC().UnixNano())
	seed := fmt.Sprintf("%x", rand.Int63n(4294967295+1))
	return seed
}
