package wesnoth

import (
	"go-wml"
)

//for era, mod, etc
//optsTag is contents of [options]
func GetModOptions (optsTag *wml.Data) map[string]string {
	options := map[string]string{}
	checkboxes := optsTag.GetTags("checkbox")
	sliders := optsTag.GetTags("slider")
	entries := optsTag.GetTags("entry")
	full := checkboxes
	full = append (full, sliders...)
	full = append (full, entries...)
	for _, opt := range full {
		id := opt.GetAttr ("id")
		if id == "" {
			continue
		}
		val := opt.GetAttr ("default")
		options[id] = val
	}
	return options
}
