package main

import (
	"strings"
	"encoding/json"
	"os"
	"fmt"
	"bytes"
	"net/http"
	
	c "wap/config"
	"github.com/go-pg/pg/v10"
)

type Player struct {
	Name string `pg:",pk"`
	Rating int
	Banned bool `pg:",notnull,use_zero"`
}

//usage bot webhook path_to_config
func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println ("usage: bot webhook path_to_config")
		os.Exit(1)
	}
	config := c.LoadConfig (args[1])
	db := pg.Connect(&pg.Options{
		User: config.DBUser,
		Database: config.DBName,
		Password: config.DBPass,
	})
	defer db.Close()
	var players []Player
	err := db.Model (&players).Where("Banned = ?", false).Order ("rating DESC").Limit(10).Select()
	if err != nil {
		panic (err)
		os.Exit(3)
	}
	var b strings.Builder
	b.WriteString ("\n")
	for i, pl := range players {
		b.WriteString (fmt.Sprintf ("%d. %s: %d\n", i+1, pl.Name, pl.Rating))
	}
	mapData := map[string]string{"content": b.String()}
	data, err := json.Marshal (mapData)
	if err != nil {
		fmt.Println ("Some crap: %s", err)
		os.Exit(4)
	}
	_, err = http.Post(args[0], "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Println ("Http error: %s", err)
		os.Exit(5)
	}
	fmt.Println ("done")
}
