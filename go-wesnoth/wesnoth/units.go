package wesnoth

import (
	"go-wml"
	"fmt"
)

func FetchUnits (path string, prep Preprocessor) (map[string]*wml.Data, error) {
	bytes, err := prep.Preprocess (path, []string{})
	if err != nil {
		return nil, err
	}
	unitsData := wml.ParseData (bytes)
	unitsTag := (unitsData.GetTags("units"))[0]
	unitTypeTages := unitsTag.GetTags ("unit_type")
	ans := map[string]*wml.Data{}
	fmt.Println(len(unitTypeTages))
	for _, u := range unitTypeTages {
		id := u.GetAttr ("id")
		
		if id == "" {
			continue
		}
		
		ans[id] = u
	}
	return ans, nil
}
