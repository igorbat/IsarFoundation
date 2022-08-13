/*// This file is part of Fastbot.
//
// Fastbot is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Fastbot is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Fastbot.  If not, see <https://www.gnu.org/licenses/>.

// fastbot project main.go
package main

import (
	"wap/config"
	"wap/server"
	"fmt"
	"encoding/json"
	"sort"
	"go-wesnoth/addon"
	"time"
	"strconv"
	"go-wml"
	"io/ioutil"
	"strings"
	"wap/types"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var (
	ratings map[string]int
	games []*Game
	banned map[string]bool
)

func IsSuperAdmin (pl string) bool {
	return config.Admins[0] == strings.ToLower (pl)
}

func IsAdmin (pl string) bool{
	for _, adm := range config.Admins {
		if strings.ToLower (pl) == adm {
			return true
		}
	}
	return false
}

func Banned (pl string) bool{
	return banned[strings.ToLower (pl)]
}

func RegisterGame (pls []string) int{
	id := len(games)
	game := &Game{Id:id, Time: time.Now().Unix(), Players: pls, RateAdjusts: []int{}}
	games = append (games, game)
	fmt.Println ("Registered game:", game)
	return id
}

func GetRating (pl string) int{
	rat, ok := ratings[strings.ToLower (pl)]
	if !ok {
		return 1500
	}
	return rat
}

func GetInfoAboutPlayer (pl string) string {
	ans := ""
	nick := strings.ToLower (pl)
	rating := GetRating(nick)
	
	played := 0
	won := 0
	lost := 0
	unfinished := []string{}
	contested := []string{}
	for _, game := range games {
		if !game.Canceled && game.HasPlayed (nick) {
			played++
			if game.HasWon (nick) {
				won++
			}
			if game.HasLost (nick) {
				lost++
			}
			if game.TeamWon == 0 {
				unfinished = append (unfinished, fmt.Sprintf ("%d", game.Id))
			}
			if game.Contested {
				contested = append (contested, fmt.Sprintf ("%d", game.Id))
			}
		}
	}
	var gamesStr string
	if played == 0 {
		gamesStr = "no games"
	} else {
		gamesStr = fmt.Sprintf ("%d games", played)
	}
	ans = fmt.Sprintf ("%s: rating %v (%s)", nick, rating, gamesStr)
	if played != 0 && played != len(unfinished) {
		nicks := []string{}
		for nickk := range ratings {
			nicks = append (nicks, nickk)
		}
		sort.Slice (nicks, func (i, j int) bool{
			return ratings[nicks[i]] > ratings[nicks[j]]
		})
		for i, v := range nicks {
			if v == nick {
				ans += fmt.Sprintf ("\nPlace in the rating: %d", i + 1)
				break
			}
		}
	}
	ans += fmt.Sprintf ("\nWon %d, lost %d, unfinished %d", won, lost,  len(unfinished))
	if len(unfinished) > 0 {
		ans += fmt.Sprintf ("\nUnfinished game ids: %s", strings.Join (unfinished, ", "))
	}
	if len(contested) > 0 {
		ans += fmt.Sprintf ("\nContested games with this player (ids): %s", strings.Join (contested, ", "))
	}
	if Banned (nick) {
		ans += fmt.Sprintf ("\n%s banned", nick)
	}
	return ans
}

func Dump (print bool) {
	ratData, _ := json.Marshal (ratings)
	gameData, _ := json.Marshal (games)
	banData, _ := json.Marshal (banned)
	if print {
		fmt.Println (string(ratData), string (gameData), string (banData))
	}
	ioutil.WriteFile("ratings.json", ratData, 0644)
	ioutil.WriteFile("games.json", gameData, 0644)
	ioutil.WriteFile("banned.json", banData, 0644)

}

func LoadDump () {
	ratData, err := ioutil.ReadFile("ratings.json")
	check(err)
	gameData, err := ioutil.ReadFile("games.json")
	check(err)
	banData, err := ioutil.ReadFile("banned.json")
	check(err)
	err = json.Unmarshal (gameData, &games)
	check(err)
	err = json.Unmarshal (ratData, &ratings)
	check(err)
	err = json.Unmarshal (banData, &banned)
	check(err)
}

func GetGame (gameId int) *Game {
	if gameId < 0 || gameId >= len(games) {
		return nil
	}
	return games[gameId]
}

func main() {
	LoadDump()
	//scenario := scenario.FromPath("afterlife_survival", "/home/user/.config/wesnoth-1.14/data/add-ons/afterlife_scenario/_main.cfg", []string{})
	//scenario := scenario.FromPath("WL_Cold_War", "/home/user/.config/wesnoth-1.14/data/add-ons/WL_Mappack/_main.cfg", []string{})
	//scenario := scenario.FromPath("multiplayer_Basilisk", "/usr/share/games/wesnoth/1.14/data/multiplayer/scenarios/2p_Caves_of_the_Basilisk.cfg", []string{})
	/*scenario := scenario.FromPath("multiplayer_Isars_Cross", "/usr/share/games/wesnoth/1.14/data/multiplayer/scenarios/4p_Isars_Cross.cfg", []string{})
	era := e.Parse ("era_default", "/usr/share/games/wesnoth/1.14/data/multiplayer/eras.cfg")
	//era := e.Parse ("The_Great_Steppe_Era", "/home/user/.config/wesnoth-1.14/data/add-ons/1The_Great_Steppe_Era/_main.cfg")
	//mod2 := m.Parse ("Rav_XP_Mod", "/home/user/.config/wesnoth-1.14/data/add-ons/XP_Modification/_main.cfg")
	mods := []m.Mod{m.Parse ("plan_unit_advance", "/home/user/.config/wesnoth-1.14/data/add-ons/pick_advance/_main.cfg")}//,
	       /*m.Parse ("Rav_XP_Mod", "/home/user/.config/wesnoth-1.14/data/add-ons/XP_Modification/_main.cfg")*/
	/*units := wesnoth.FetchUnits ("/usr/share/games/wesnoth/1.14/data/core/units.cfg")
	//units := wesnoth.FetchUnits ("/home/user/.config/wesnoth-1.14/data/add-ons/1The_Great_Steppe_Era/_main.cfg")*/
	/*s := server.NewServer(
		config.Hostname,
		config.Port,
		config.Version,
		config.Username,
		config.Password,
		config.Timer.Enabled,
		config.Timer.InitTime,
		config.Timer.TurnBonus,
		config.Timer.ReservoirTime,
		config.Timer.ActionBonus,
		config.Timeout,
		false,
		)
	fmt.Println("Log in started")
	err := s.Connect()
	check(err)
	for true {
		fmt.Println("Isar hosted")
		time.Sleep(time.Second * 1)
		s.HostGame(config.Scenario, config.Era, config.Mods, []addon.Addon{}, fmt.Sprintf ("Isar Foundation Game #%d!", len (games)), "")
		cycle := true
		//srv.Listen()
		for cycle{
			data, err := s.GetServerInput()
			if err != nil {
				break
			}
			switch {
				case data.Contains("name") && data.Contains("side") && s.Sides.FreeSlots() > 0:
					//fmt.Println ("HHUII")
					name := data["name"].(string)
					side, _ := strconv.Atoi(data["side"].(string))
					fmt.Printf("%s wants side %d\n", name, side)
					// if not blacklisted
					if s.Sides.HasSide(side) && !s.Sides.HasPlayer(name) && !Banned (name){
						s.SetSidePlayer (side, name, true)
						s.Message ("Welcome to the Isar Foundation, a place for rated Isar and fun! I'll start the game immediately when all slots are filled")
						time.Sleep (time.Millisecond * 300)
						//s.Message ("I host from Dwarftough nickname temporarily for technical test purposes, IsarFoundation nick will be back soon :)")
						if s.Sides.MustStart() {
							s.StartGame(config.Era, config.Units, true)
							players := make([]string, len(s.Sides))
							for i, val := range s.Sides {
								players[i] = strings.ToLower (val.Player)
							}
							id := RegisterGame(players)
							welcomeMes := "Ratings:"
							for _, pl := range players {
								welcomeMes += fmt.Sprintf ("\n%s: %v", pl, GetRating (pl))
							}
							s.InGameMessage (welcomeMes)
							s.InGameMessage (fmt.Sprintf("Winners: don't forget to report whispering 'won %d' to me. Beware of timer (3m30s) and have fun!", id))
							s.LeaveGame()
							Dump(false)
							cycle = false
							break
						}
					} else if name != s.Username && !s.Observers.ContainsValue(name) {
						s.Observers = append(s.Observers, name)
						if Banned (name) {
							s.Whisper (name, fmt.Sprintf ("Sorry, %s, you are banned, contact the admins for details", name))
						}
					}
				case data.Contains("side_drop"):
					side_drop := data["side_drop"].(wml.Data)
					if side_drop.Contains("side_num") {
						side, _ := strconv.Atoi(side_drop["side_num"].(string))
						s.ClearSide(side)
					}
				case data.Contains("observer"):
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
					}
				case data.Contains("leave_game"):
					for _, v := range s.Sides {
						v.Player = ""
						v.Ready = false
					}
					s.InGame = false
					cycle = false*/
					//for _, v := range s.Sides {
					//	s.ChangeSide(v.Side, "insert", wml.Data{"color": v.Color})
					//}
					/*
				case data.Contains("whisper"):
					whisper := data["whisper"].(wml.Data)
					if whisper.Contains("message") && whisper.Contains("receiver") && whisper.Contains("sender") {
						text := whisper["message"].(string)
						receiver := whisper["receiver"].(string)
						sender := whisper["sender"].(string)
						fmt.Printf("%s whispers %s\n", sender, text)
						time.Sleep (time.Second * 1)
						if receiver == s.Username && IsAdmin(sender) {
							//admin commands
							command := strings.Fields(text)
							if len(command) > 0 {
								switch {
									case command[0] == "admins" && len(command) == 1:
										s.Whisper(sender, "Admin list: "+strings.Join(config.Admins, ", "))
									case command[0] == "contested" && len(command) == 1:
										contested := []string{}
										for _, game := range games {
											if !game.Canceled && game.Contested {
												contested = append (contested, fmt.Sprintf("%d", game.Id))
											}
										}
										s.Whisper (sender, "Contested games: "+strings.Join(contested, ", "))
									case command[0] == "unfinished" && len(command) == 1:
										contested := []string{}
										for _, game := range games {
											if !game.Canceled && game.TeamWon == 0 {
												contested = append (contested, fmt.Sprintf("%d", game.Id))
											}
										}
										s.Whisper (sender, "Unfinished games: "+strings.Join(contested, ", "))
									case command[0] == "cancel" && len(command) == 2:
										gameId := types.ParseInt(command[1], -1)
										game := GetGame (gameId)
										if game != nil {
											game.UndoRatings()
											game.Canceled = true
											s.Whisper(sender, "Game canceled")
											Dump (false)
										}
									case command[0] == "undo_report" && len(command) == 2:
										gameId := types.ParseInt(command[1], -1)
										game := GetGame (gameId)
										if game != nil {
											if game.TeamWon != 0 {
												game.UndoRatings()
												s.Whisper(sender, "Game unreported")
												Dump (false)
											} else {
												s.Whisper (sender, "Game isn't reported, nothing to undo")
											}
										} else {
											s.Whisper(sender, "No such game")
										}
									case (command[0] == "force_report") && len(command) == 3:
										gameId := types.ParseInt(command[1], -1)
										winner := command[2]
										game := GetGame (gameId)
										if game != nil {
											if game.HasPlayed (winner) {
												if game.TeamWon != 0 {
													//undo ratings
													game.UndoRatings()
												}
												game.ReportGame (winner)
												game.Reporter = "admin " + sender //admin as reporter
												s.Whisper(sender, "Game force-reported")
												Dump (false)
											} else {
												s.Whisper(sender, fmt.Sprintf("%s hasn't played this game", winner))
											}
										} else {
											s.Whisper(sender, "Wrong id of the game")
										}*/
									//case command[0] == "to_check" && len(command) == 1:
							//msg := s.Ladder.ShowContested()
							//s.Whisper(sender, "Contested List: " + msg)
						/*case command[0] =="to_finish" && len(command) == 1:
							msg := s.Ladder.ShowUnfinished()
							s.Whisper(sender, "Unfinished List: " + msg)*//*
									case command[0] == "ban" && len(command) == 2:
										banned[strings.ToLower (command[1])] = true
										s.Whisper(sender, "Banned")
									case command[0] == "unban" && len(command) == 2:
										banned[strings.ToLower (command[1])] = false
										s.Whisper(sender, "Unbanned")
									case command[0] == "uncontest" && len(command) == 2:
										gameId := types.ParseInt(command[1], -1)
										game := GetGame (gameId)
										if game != nil {
											if game.Contested {
												game.Contested = false
												game.ContestedBy = ""
												s.Whisper (sender, "Game uncontested")
											} else {
												s.Whisper(sender, "Game isn't contested")
											}
										} else {
											s.Whisper(sender, "Wrong id of the game")
										}*/
									/*case command[0] == "force_report" && len(command) == 3:
										gameId := types.ParseInt(command[1], -1)
										msg := s.Ladder.GameReported(command[2], gameId)
										s.Whisper(sender, msg)
									case command[0] == "force_contest" && len(command) == 3:
										gameId := types.ParseInt(command[1], -1)
										msg := s.Ladder.GameContested(command[2], gameId)
										s.Whisper(sender, msg)
									case command[0] == "uncontest" && len(command) == 3:
										gameId := types.ParseInt(command[1], -1)
										msg := s.Ladder.GameUnContested(command[2], gameId)
										s.Whisper(sender, msg)
									case command[0] == "force_finish" && len(command) == 1:
										s.ForceFinish = true
										s.Whisper(sender, "Logging out totally...")
										s.LeaveGame()
										s.Disconnect()*/
									/*case command[0] == "stop" && len(command) == 1:
										if IsSuperAdmin (sender) {
											s.Whisper(sender, "Logging out...")
											s.LeaveGame()
											Dump (true)
											panic ("Force quit from super admin "+ sender)
										}
									case command[0] == "dump" && len(command) == 1:
										Dump (true)
										s.Whisper(sender, "Dump requested")
									case command[0] == "load" && len(command) == 1:
										if IsSuperAdmin (sender) {
											LoadDump ()
											s.Whisper(sender, "Reread my state from the dump")
										}
									case command[0] == "admin_help" && len(command) == 1:
										s.Whisper(sender, "Command list:\nadmins, contested, ban, unban, dump, load, cancel, undo_report, stop, unfinished")
									}
							}
						}
						// for rabotyagi  //////////////////////////////////////////////////////////////////////////////
						if receiver == s.Username {
						command := strings.Fields(text)
						if len(command) > 0 {
							switch {
							case (command[0] == "won" || command[0] == "win") && len(command) == 2:
								if Banned (sender) {
									s.Whisper(sender, "You are banned, contact admins")
									break
								}
								gameId := types.ParseInt(command[1], -1)
								game := GetGame (gameId)
								if game != nil {
									if game.HasPlayed (sender) {
										if game.TeamWon == 0 {
											game.ReportGame (sender)
											s.Whisper(sender, "Game reported")
											Dump (false)
										} else {
											if game.HasWon (sender) {
												s.Whisper(sender, "Game already reported as your win")
											} else {
												s.Whisper(sender, fmt.Sprintf("Game already reported as your loss, you may contest it if it's wrong, type 'contest %d'", gameId))
											}
										}
									} else {
										s.Whisper(sender, "You haven't played this game")
									}
								} else {
									s.Whisper(sender, "Wrong id of the game")
								}
							case command[0] == "playerinfo" && len(command) == 2:
								s.Whisper(sender, GetInfoAboutPlayer (command[1]))
							case command[0] == "leaderboard" && len(command) == 1:
								nicks := []string{}
								for nick := range ratings {
									nicks = append (nicks, nick)
								}
								sort.Slice (nicks, func (i, j int) bool{
									return ratings[nicks[i]] > ratings[nicks[j]]
								})
								var b strings.Builder
								for i:=0; i < len (nicks) && i < 6; i = i + 1 {
									b.WriteString (fmt.Sprintf ("%s: %d\n", nicks[i], ratings[nicks[i]]))
								}
								s.Whisper (sender, b.String())
							case command[0] == "me" && len(command) == 1:
								s.Whisper(sender, GetInfoAboutPlayer (sender))
							
							case command[0] == "contest" && len(command) == 2:
								gameId := types.ParseInt(command[1], -1)
								game := GetGame (gameId)
								if game != nil {
									if game.HasPlayed (sender) {
										if game.Contested {
											s.Whisper (sender, "Game has already been contested")
										} else {
											game.Contested = true
											game.ContestedBy = sender
											s.Whisper (sender, "Game contested")
											Dump (false)
										}
									} else {
										s.Whisper(sender, "You haven't played this game")
									}
								} else {
									s.Whisper(sender, "Wrong id of the game")
								}
							
							case command[0] == "gameinfo" && len(command) == 2:
								gameId := types.ParseInt(command[1], -1)
								game := GetGame (gameId)
								if game != nil {
									s.Whisper (sender, "Players: "+ strings.Join (game.Players, ", ")+"\nStarted: "+time.Unix (game.Time,0).UTC().Format(time.UnixDate))
									if game.Canceled {
										s.Whisper(sender, "Game canceled")
									} else {
										if game.TeamWon == 0 {
											s.Whisper (sender, "Still being played or unreported")
										} else {
											winners := []string{}
											for _, p := range game.Players {
												if game.HasWon (p) {
													winners = append (winners, p)
												}
											}
											s.Whisper(sender, "Winners: "+strings.Join (winners, ", "))
											s.Whisper(sender, "Reported by "+game.Reporter)
										}
										if game.Contested {
											s.Whisper(sender, "Contested by "+game.ContestedBy)
										}
									}
								} else {
									s.Whisper(sender, "Wrong id of the game")
								}
							default:
								s.Whisper(sender, "Command list:\n"+
									"won <game_id> - report the game you have played\n" +
									"contest <game_id> - contest the game you have played\n" +
									"gameinfo <game_id> - info about the game\n" +
									"leaderboard - top 6 players\n"+
									"playerinfo <nickname> - info about the player\n" +
									"me - info about you\n" +
									"help - request command reference")
							}
						}
					}
				}
			}
	}
		time.Sleep (time.Second * 10)
		if s.ForceFinish {
			break
		}
	}
}*/
