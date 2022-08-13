// This file is part of Go WML.
//
// Go WML is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Go WML is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Go WML.  If not, see <https://www.gnu.org/licenses/>.

package wml

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type AttrMap map[string]interface{}

type Data struct {
	Attrs AttrMap
	Tags []*Tag
}

func NewData () *Data{
	return &Data {
		Attrs: map[string]interface{}{},
		Tags: []*Tag{},
	}
}

func NewDataAttrs (attrs AttrMap) *Data {
	return &Data {
		Attrs: attrs,
		Tags: []*Tag{},
	}
}

func NewDataAttrsTags (attrs AttrMap, tags ...*Tag) *Data {
	return &Data {
		Attrs: attrs,
		Tags: tags,
	}
}

func NewDataTags (tags ...*Tag) *Data {
	return &Data {
		Attrs: map[string]interface{}{},
		Tags: tags,
	}
}

type RawData string
//type Multiple []interface{}
type Tr string
type Domain struct {
	V interface{}
	D string
}

// Merges multiple Datas, priority to the last parameter
func MergeData(first *Data, others ...*Data) *Data {
	data := NewData ()
	for k, v := range first.Attrs {
		data.AddAttr(k, v)
	}
	for _, d := range others {
		for k, v := range d.Attrs {
			data.AddAttr(k, v)
		}
	}
	for _, tag := range first.Tags {
		data.Tags = append (data.Tags, tag)
	}
	for _, d := range others {
		for _, tag := range d.Tags {
			data.Tags = append (data.Tags, tag)
		}
	}
	return data
}

func (d *Data) Bytes() []byte {
	return []byte(d.String())
}

func (d *Data) String() string {
	return d.Indent(0)
}

func (d *Data) Contains(key string) bool {
	_, ok := d.Attrs[key]
	return ok
}

func (d *Data) ContainsTag(key string) bool {
	for _, tag := range d.Tags {
		if tag.Name == key {
			return true
		}
	}
	return false
}

/*func (d *Data) Single() (string, interface{}, error) {
	if len(*d) == 1 {
		for k, v := range *d {
			return k, v, nil
		}
	}
	return "", nil, errors.New("The data isn't single-valued.")
}*/

/*func (d *Data) ToTag() (Tag, error) {
	key, value, err := d.Single()
	if err == nil {
		if data, ok := value.(Data); ok {
			return Tag{key, data}, nil
		}
	}
	return Tag{}, errors.New("The data isn't a tag.")
}*/
/*
func (d *Data) ReadData(path string) (Data, error) {
	value, err := d.Read(path)
	if err != nil {
		return nil, err
	}
	if data, ok := value.(Data); ok {
		return data, nil
	} else {
		return nil, errors.New("Incorrect type of the Data attribute.")
	}
}

func (d *Data) ReadString(path string) (string, error) {
	value, err := d.Read(path)
	if err != nil {
		return "", err
	}
	if data, ok := value.(string); ok {
		return data, nil
	} else {
		return "", errors.New("Incorrect type of the Data attribute.")
	}
}*/

// Read nested data using dot notation.
/*func (d *Data) Read(path string) (interface{}, error) {
	pathSlice := strings.Split(path, ".")

	var data Data = *d
	for _, v := range pathSlice[:len(pathSlice)-1] {
		if data.Contains(v) {
			switch data[v].(type) {
			case Data:
				data = data[v].(Data)
			default:
				return nil, errors.New("The path doesn't exist")
			}
		}
	}

	key := pathSlice[len(pathSlice)-1]
	if data.Contains(key) {
		return data[key], nil
	} else {
		return nil, errors.New("The path doesn't exist")
	}
}
*/
func quotes (in string) string {
	return strings.ReplaceAll (in, `"`, `""`)
}

func (d *Data) Indent(nesting uint) string {
	tabulation := strings.Repeat("\t", int(nesting))
	var keys []string
	for k := range d.Attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	attributes := ""
	subTags := ""
	for _, key := range keys {
		var prepend string
		var value interface{}
		// Check whether the data is Domain or not. If it's Domain, add a textdomain line.
		switch (d.Attrs)[key].(type) {
		case Domain:
			prepend = tabulation + "#textdomain " + (d.Attrs)[key].(Domain).D + "\n"
			value = (d.Attrs)[key].(Domain).V
		default:
			prepend = ""
			value = (d.Attrs)[key]
		}
		switch value.(type) {
		case bool:
			var v string
			if value.(bool) {
				v = "yes"
			} else {
				v = "no"
			}
			attributes += prepend
			attributes += tabulation + key + "=" + v + "\n"
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			attributes += prepend
			attributes += tabulation + key + "=" + fmt.Sprintf("%v", value) + "\n"
		case string:
			attributes += prepend
			attributes += tabulation + key + "=\"" + quotes(value.(string)) + "\"\n"
		case Tr:
			attributes += prepend
			attributes += tabulation + key + "=_\"" + quotes(string(value.(Tr))) + "\"\n"
		default: 
			panic ("Unknown type "+fmt.Sprintf ("%T", value))
		}
	}
	for _, tag := range d.Tags {
		subTags += tabulation + "\n"
		subTags += tag.Indent (nesting)
	}
	return attributes + subTags
}

func (d *Data) GetTags (name string) (ans []*Data) {
	ans = []*Data{}
	for _, tag := range d.Tags {
		if tag.Name == name {
			ans = append (ans, tag.Data)
		}
	}
	return
}

func (d *Data) GetTag (name string) (*Data, bool) {
	for _, tag := range d.Tags {
		if tag.Name == name {
			return tag.Data, true
		}
	}
	return nil, false
}

func (d *Data) GetAttr (name string) string {
	val, ok := d.Attrs[name]
	if !ok {
		return ""
	}
	switch val.(type) {
		case Tr:
			return string(val.(Tr))
		case string:
			return val.(string)
		default:
			return ""
	}
}

func (d *Data) GetAttrInt (name string) (int, error) {
	val, ok := d.Attrs[name]
	if !ok {
		return 0, errors.New("No such attr")
	}
	switch val.(type) {
		case int:
			return val.(int), nil
		case string:
			str := val.(string)
			num, err := strconv.Atoi (str)
			return num, err
		default:
			return 0, errors.New("Not a number")
	}
}

func (d *Data) AddTagByName (name string, data *Data){
	d.Tags = append (d.Tags, &Tag {Name: name, Data: data})
}

func (d *Data) AddTag (tag *Tag){
	d.Tags = append (d.Tags, tag)
}

func (d *Data) AddAttr (name string, val interface{}){
	d.Attrs[name] = val
}
