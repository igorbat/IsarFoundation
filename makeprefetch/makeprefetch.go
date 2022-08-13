package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"fmt"
)

var (
	Wesnoth = "/usr/games/wesnoth"
	WesnothData = "/usr/share/games/wesnoth/1.14/data/"
	Output  = os.TempDir() + "/go-wesnoth/output"
)

func Preprocess(filePath string, defines []string) ([]byte, error) {
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
		return nil, err
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
		return nil, err
	}
	return result, nil
}

func main() {
	args := os.Args[1:]
	if len(args) < 3 {
		fmt.Println("Usage: makeprefetch <path to wesnoth executable> <path to wesnoth data> <file to prefetch> [optional defines]")
		return
	}
	Wesnoth = args[0]
	WesnothData = args[1]
	data, err := Preprocess(args[2], args[3:])
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
