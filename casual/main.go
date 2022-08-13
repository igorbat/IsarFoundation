package main

import (
	"fmt"
	"time"
	"os"
	
	c "wap/config"
	"wap/server"
	"go-wesnoth/mod"
	r "go-wesnoth/resource"
	e "go-wesnoth/era"
	"go-wesnoth/wesnoth"
	"go-wesnoth/scenario"
	"go-wesnoth/game"
	"go-wml"
	"runtime"
	"strconv"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Watchdog struct {
    interval time.Duration
    timer *time.Timer
}

func NewWatchdog(interval time.Duration, callback func()) *Watchdog {
    w := Watchdog{
        interval: interval,
        timer: time.AfterFunc(interval, callback),
    }
    return &w
}

func (w *Watchdog) Stop() {
    w.timer.Stop()
}

func (w *Watchdog) Kick() {
    w.timer.Stop()
    w.timer.Reset(w.interval)
}

func main () {
	wdog := NewWatchdog (15 * time.Minute, func(){
			panic("Watchdog timer expired!")
		})
	const wesnothVer = "1.16.0"
	config := c.LoadConfig ("config.json")
	var prep wesnoth.Preprocessor
    	if len(os.Args) > 1 {
		prep = &wesnoth.PrefetchPreprocessor{}
	} else {
		prep = &wesnoth.WesnothPreprocessor{
			Wesnoth: config.WesnothBinary,
			WesnothData: config.WesnothData,
		}
	}
	
	res := map[string]r.Resource{}
	for _, resPath := range config.ResPaths {
		resMap := r.Parse (resPath, prep)
		for id, rrr := range resMap {
			res[id] = rrr
		}
	}
	era := e.ParseWithResources (config.EraId, config.EraPath, prep, res)
	mods := []mod.Mod{}
	for mId, mPath := range config.ModPaths {
		mods = append (mods, mod.Parse(mId, mPath, prep))
	}
	
	fmt.Println(config.ScenarioPath)
	
	units, err := wesnoth.FetchUnits (config.UnitsPath, prep)
	check(err)
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
	err = s.ConnectEnhanced(!config.NoTLS)
	check(err)
	
	for true {
		//fmt.Println("Isar hosted")
		time.Sleep(time.Second * 1)
		s.HostGameFromTemplate(sc, g, config.GameTitle, "")
		_ = GameListen (s, config.GreetMessage, config.ExtraMessage, config.UnqualifiedMessage, nil, era, units, wdog) 
		time.Sleep (time.Millisecond * 500)
		if s.ForceFinish {
			break
		}
	}
}

func GameListen (s *server.Server, greetMsg, extraMsg, unqualifMsg string, shuffler server.Shuffler, era e.Era,units map[string]*wml.Data, wdog *Watchdog) error {
	defer runtime.GC()
	for true{
	data, err := s.GetServerInput(1024)//we don't need more than 1KB answers to parse
	if err != nil {
		return err
	}
	wdog.Kick()
	switch {
		case data.Contains("name") && data.Contains("side") && s.Sides.FreeSlots() > 0:
			name := data.GetAttr("name")
			side, _ := strconv.Atoi(data.GetAttr("side"))
			fmt.Printf("%s wants side %d\n", name, side)
			for _, val := range s.Sides {
				fmt.Println (val.Side, val.Player)
			}
			// if not blacklisted
			if s.Sides.HasSide(side) && !s.Sides.HasPlayer(name){
				s.SetSidePlayer (side, name, true)
				s.Message (greetMsg) //"Welcome to the Isar Foundation, a place for rated Isar and fun! I'll start the game immediately when all slots are filled")
				time.Sleep (time.Millisecond * 300)
				if s.Sides.MustStart() {
					_, err := s.StartGameEx(era, units, true, shuffler)
					if err != nil {
						return err
					}
					players := []string{}
					for _, val := range s.Sides {
						if val.Controller == "human" {
							players = append (players, strings.ToLower (val.Player))
						}
					}
					s.InGameMessage (extraMsg)//("Our discord: https://discord.gg/AmyzYNXrnc")
					s.InGameMessage ("It's an unrated game. Good luck and have fun!")
					s.LeaveGame()
					return nil
				}
			} else if name != s.Username && !s.Observers.ContainsValue(name) {
				s.Observers = append(s.Observers, name)
			}
		case data.ContainsTag("side_drop"):
			side_drop, _ := data.GetTag("side_drop")
			if side_drop.Contains("side_num") {
				side, _ := strconv.Atoi(side_drop.GetAttr("side_num"))
				fmt.Printf("%d side dropped\n", side)
				s.ClearSide(side)
			}
		/*case data.ContainsTag("observer"):
			observer := data["observer"].(wml.Data)
			if observer.Contains("name") {
				name := observer["name"].(string)
				if name != s.Username && !s.Observers.ContainsValue(name) {
					s.Observers = append(s.Observers, name)
				}
			}
		case data.Contains("observer_quit"):
			ObserverQuit := data["observer_quit"].(wml.Data)
			if ObserverQuit.Contains("name") {
				name := ObserverQuit["name"].(string)
				if name != s.Username && s.Observers.ContainsValue(name) {
					s.Observers.DeleteValue(name)
				}
			}*/
		case data.ContainsTag("leave_game"):
			for _, v := range s.Sides {
				v.Player = ""
				v.Ready = false
			}
			s.InGame = false
			return nil
		case data.ContainsTag("whisper"):
			whisper, _ := data.GetTag("whisper")
			if !whisper.Contains("message") || !whisper.Contains("receiver") || !whisper.Contains("sender") {
				return nil
			}
			sender := whisper.GetAttr("sender")
			s.Whisper(sender, "I'm a bot for unrated casual Isar, no commands available :)")
			
	}
	}
	return nil				
}
