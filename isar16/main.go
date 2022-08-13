package main

import (
	"fmt"
	"time"
	"os"
	
	"newwesbot"
	c "wap/config"
	"wap/server"
	"go-wesnoth/mod"
	e "go-wesnoth/era"
	"go-wesnoth/wesnoth"
	"go-wesnoth/scenario"
	"go-wesnoth/game"
	//"newladder/glicko"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main () {
	const wesnothVer = "1.16.0"
	config := c.LoadConfig ("config.json")
	db := pg.Connect(&pg.Options{
		User: config.DBUser,
		Database: config.DBName,
		Password: config.DBPass,
	})
	defer db.Close()
	if len(os.Args) == 2 && os.Args[1] == "bootstrap" {
		models := []interface{}{
			(*newwesbot.Player)(nil),
			(*newwesbot.Game)(nil),
		}

		for _, model := range models {
			err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			    
			})
			if err != nil {
				panic(err)
			}
		}
		return
	}
    	var prep wesnoth.Preprocessor
    	if len(os.Args) > 1 {
		prep = &wesnoth.PrefetchPreprocessor{}
	} else {
		prep = &wesnoth.WesnothPreprocessor{
			Wesnoth: config.WesnothBinary,
			WesnothData: config.WesnothData,
		}
	}
	
	lad := newwesbot.NewGenericLadder (db, config.Admins, IsarParams{}, true)
	bot := newwesbot.NewBot (lad)
	newwesbot.AddDefaultsToBot (bot)
	
	era := e.Parse (config.EraId, config.EraPath, prep)
	mods := []mod.Mod{}
	for mId, mPath := range config.ModPaths {
		mods = append (mods, mod.Parse(mId, mPath, prep))
	}
	
	fmt.Println(config.ScenarioPath)
	
	units, errr := wesnoth.FetchUnits (config.UnitsPath, prep)
	check(errr)
	sc := scenario.FromPath(config.ScenarioId, config.ScenarioPath, []string{}, prep)
	s := server.NewServer(
		config.Hostname,
		config.Port,
		wesnothVer,
		config.Username,
		config.Password,
		config.Timer.Enabled,
		config.Timer.InitTime,
		config.Timer.TurnBonus,
		config.Timer.ReservoirTime,
		config.Timer.ActionBonus,
		time.Second * 30,
		false,
		)
	g := game.NewGame("",
		sc,
		era, mods, config.Addons,
		s.TimerEnabled, s.InitTime, s.TurnBonus, s.ReservoirTime, s.ActionBonus,
		wesnothVer)
	fmt.Println("Log in started")
	err := s.ConnectEnhanced(!config.NoTLS)
	check(err)
	
	for true {
		//fmt.Println("Isar hosted")
		time.Sleep(time.Second * 1)
		s.HostGameFromTemplate(sc, g, fmt.Sprintf ("%s #%d!", config.GameTitle, lad.NextGameId()), "")
		_ = bot.GameListen (s, config.GreetMessage, config.ExtraMessage, config.UnqualifiedMessage, nil, era, units) 
		time.Sleep (time.Millisecond * 500)
		if s.ForceFinish {
			break
		}
	}
}
