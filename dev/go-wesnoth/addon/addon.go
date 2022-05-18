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
