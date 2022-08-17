package newwesbot

import (
	"strings"
	"wap/server"
	e "go-wesnoth/era"
	"strconv"
	"fmt"
	"time"
	"go-wml"
	"runtime"
)

type Command func (Ladder, *server.Server, string, []string)
type BotCommand struct{
	Comm Command
	Name string
	NumArgs int
	Admin bool
}

func NewBotCommand (name string, num int, admin bool, comm Command) BotCommand {
	if num < 0 {
		panic ("Negative numargs for command "+name)
	}
	return BotCommand {
		Comm: comm,
		Name: strings.ToLower (name),
		Admin: admin,
		NumArgs: num,
	}
}

type Bot struct {
	admComms map[string]BotCommand
	comms map[string]BotCommand
	lad Ladder
	wdog *Watchdog
	fixFactions func (int, []string)
}

func NewBot (l Ladder) *Bot {
	return NewBotFactionsFix (l, nil)
}

func NewBotFactionsFix (l Ladder, factionFixator func (int, []string)) *Bot {
	return &Bot {
		admComms: map[string]BotCommand{},
		comms: map[string]BotCommand{},
		lad: l,
		wdog: NewWatchdog (15 * time.Minute, func(){
			panic("Watchdog timer expired!")
		}),
		fixFactions: factionFixator,
	}
}

func (b *Bot) AddCommand (c BotCommand) {
	_, ok := b.admComms[strings.ToLower (c.Name)]
	_, ok2 := b.comms[strings.ToLower (c.Name)]
	if ok || ok2 {
		panic ("Duplicate command "+c.Name)
	}
	if c.Admin {
		b.admComms[strings.ToLower (c.Name)] = c
	} else {
		b.comms[strings.ToLower (c.Name)] = c
	}
}

func (b *Bot) GameListen (s *server.Server, greetMsg, extraMsg, unqualifMsg string, shuffler server.Shuffler, era e.Era,units map[string]*wml.Data) error {
	defer runtime.GC()
	for true{
	data, err := s.GetServerInput(1024)//we don't need more than 1KB answers to parse
	if err != nil {
		return err
	}
	b.wdog.Kick()
	switch {
		case data.Contains("name") && data.Contains("side") && s.Sides.FreeSlots() > 0:
			name := data.GetAttr("name")
			side, _ := strconv.Atoi(data.GetAttr("side"))
			fmt.Printf("%s wants side %d\n", name, side)
			for _, val := range s.Sides {
				fmt.Println (val.Side, val.Player)
			}
			// if not blacklisted
			if s.Sides.HasSide(side) && !s.Sides.HasPlayer(name) && !b.lad.GetPlayer (name).Banned && b.lad.IsQualified (name) {
				s.SetSidePlayer (side, name, true)
				s.Message (greetMsg) //"Welcome to the Isar Foundation, a place for rated Isar and fun! I'll start the game immediately when all slots are filled")
				time.Sleep (time.Millisecond * 300)
				if s.Sides.MustStart() {
					factions, err := s.StartGameEx(era, units, true, shuffler)
					if err != nil {
						return err
					}
					players := []string{}
					for _, val := range s.Sides {
						if val.Controller == "human" {
							players = append (players, strings.ToLower (val.Player))
						}
					}
					id := b.lad.RegisterGame(players)
					welcomeMes := "Ratings:"
					for _, pl := range players {
						welcomeMes += fmt.Sprintf ("\n%s: %d", pl, b.lad.GetPlayer (pl).Rating)
					}
					s.InGameMessage (welcomeMes)
					s.InGameMessage (fmt.Sprintf("Winners: don't forget to report typing '/m %s won %d' in the chat. Beware of timer and have fun!", s.Username, id))
					s.InGameMessage (extraMsg)//("Our discord: https://discord.gg/AmyzYNXrnc")
					if b.fixFactions != nil {
						b.fixFactions (id, factions)
					}
					s.LeaveGame()
					return nil
				}
			} else if name != s.Username && !s.Observers.ContainsValue(name) {
				s.Observers = append(s.Observers, name)
				if b.lad.GetPlayer (name).Banned {
					s.Whisper (name, fmt.Sprintf ("Sorry, %s, you are banned, contact the admins for details", name))
				} else {
					s.Whisper (name, unqualifMsg)
					s.Message (fmt.Sprintf("%s, %s", name, unqualifMsg))
				}
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
			text := whisper.GetAttr("message")
			receiver := whisper.GetAttr("receiver")
			sender := whisper.GetAttr("sender")
			fmt.Printf("%s whispers %s\n", sender, text)
			time.Sleep (time.Second * 1)
			command := strings.Fields(text)
			if len (command) == 0 {
				return nil
			}
			found := false
			if receiver == s.Username && b.lad.IsAdmin(sender) {
				comm, ok := b.admComms[strings.ToLower (command[0])]
				if ok {
					if comm.NumArgs == len(command) - 1 {
						found = true
						comm.Comm (b.lad, s, sender, command[1:])
					}
				}	
			}
			if !found && receiver == s.Username {
				comm, ok := b.comms[strings.ToLower (command[0])]
				if ok {
					if comm.NumArgs == len(command) - 1 {
						found = true
						comm.Comm (b.lad, s, sender, command[1:])
					}
				}
			}
			if !found {
				s.Whisper(sender, "Command list:\n"+
									"won <game_id> - report the game you have played\n" +
									"contest <game_id> - contest the game you have played\n" +
									"game <game_id> - info about the game\n" +
									"leaderboard - top 10 players\n"+
									"info <nickname> - info about the player\n" +
									"me - info about you\n")
			}
	}
	}
	return nil				
}
