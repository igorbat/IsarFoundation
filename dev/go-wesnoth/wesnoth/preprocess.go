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

package wesnoth

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	Wesnoth = "/usr/games/wesnoth"
	WesnothData = "/usr/share/games/wesnoth/1.14/data/"
	Output  = os.TempDir() + "/go-wesnoth/output"
	cache = map[string][]byte{}
	PrefetchedMode = false
)

func simplePrefetch (filePath string) []byte{
	result, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic ("Error when simple prefetching: "+err.Error())
	}
	return result
}

func AdvancedPreprocess(filePath string, defines []string) []byte {
	if PrefetchedMode {
		return simplePrefetch (filePath)
	}
	defines = append(defines, "MULTIPLAYER")
	if _, err := os.Stat(Output); os.IsNotExist(err) {
		os.MkdirAll(Output, 0755)
	}
	precmd := exec.Command(
		Wesnoth,
		"-p",
		WesnothData,
		Output,
		"--preprocess-defines="+strings.Join(defines, ","),
		"--preprocess-output-macros=macros.advanced")
	err := precmd.Run()
	if err != nil {
		panic(err)
	}
	cmd := exec.Command(
		Wesnoth,
		"-p",
		filePath,
		Output,
		"--preprocess-defines="+strings.Join(defines, ","),
		"--preprocess-input-macros="+Output+"/macros.advanced")
	cmd.Run()
	result, err := ioutil.ReadFile(Output + "/" + filepath.Base(filePath))
	if err != nil {
		panic ("Error when advanced preprocessing: "+err.Error())
	}
	return result
}


func Preprocess(filePath string, defines []string) []byte {
	if PrefetchedMode {
		return simplePrefetch (filePath)
	}
	//_, present := cache[filePath]
	//if !present {
		defines = append(defines, "MULTIPLAYER")
		if _, err := os.Stat(Output); os.IsNotExist(err) {
			os.MkdirAll(Output, 0755)
		}
		cmd := exec.Command(
			Wesnoth,
			"-p",
			filePath,
			Output,
			"--preprocess-defines="+strings.Join(defines, ","),
		)
		cmd.Run()
		result, err := ioutil.ReadFile(Output + "/" + filepath.Base(filePath))
		if err != nil {
			panic ("Error when preprocessing: "+err.Error())
		}
		return result
		//cache[filePath] = result
	//}
	//return cache[filePath]
}
