package resource

import (
	"fmt"
	"go-wesnoth/wesnoth"
	"go-wml"
)

type Resource struct {
	Id       string
	Events   []*wml.Data
}


func Parse(path string, prep wesnoth.Preprocessor) map[string]Resource {
	resources := map[string]Resource{}
	content, err := prep.Preprocess(path, nil)
	if err != nil {
		panic ("Failed to prep "+err.Error())
	}
	fmt.Println("resources preprocess finished")
	
	resData := wml.ParseData (content)
	resTags := resData.GetTags ("resource")
	for _, r := range resTags {
		res := Resource{
			Id: r.GetAttr("id"),
			Events: r.GetTags ("event"),
		}
		resources[res.Id] = res
	}
	return resources
}
