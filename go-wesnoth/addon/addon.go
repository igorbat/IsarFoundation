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

package addon

import "go-wml"

type Content struct {
	Id       string
	Name     string
	Type     string
}

type Addon struct {
	Id       string
	Name     string
	Version  string
	Require  bool
	Contents []Content
}

func (c *Content) ToWML() *wml.Data{
	return wml.NewDataAttrs(wml.AttrMap{"id": c.Id, "name": c.Name, "type": c.Type})
}

func (a *Addon) ToWML () *wml.Data{
	data := wml.NewDataAttrs(wml.AttrMap{"id": a.Id, "name": a.Name, "version": a.Version, "require": a.Require})
	for _, c := range a.Contents {
		data.AddTagByName ("content", c.ToWML())
	}
	return data
}
