package newwesbot

import (
	"strings"
	"wap/server"
	"wap/types"
	"fmt"
	"time"
)

//DEFAULT GENERIC COMMANDS

//case command[0] == "admins" && len(command) == 1:
//	s.Whisper(sender, "Admin list: "+strings.Join(config.Admins, ", "))
func getInfoAboutPlayer (l Ladder, pl string) string {
	ans := ""
	nick := strings.ToLower (pl)
	player := l.GetPlayer(nick)
	
	played := 0
	won := 0
	lost := 0
	unfinished := []string{}
	contested := []string{}
	games := l.GetGamesOf (nick)
	for _, game := range games {
		played++
		if l.HasWon (game, nick) {
			won++
		}
		if l.HasLost (game,nick) {
			lost++
		}
		if game.TeamWon == 0 {
			unfinished = append (unfinished, fmt.Sprintf ("%d", game.Id))
		}
		if game.Contested {
			contested = append (contested, fmt.Sprintf ("%d", game.Id))
		}
	}
	var gamesStr string
	if played == 0 {
		gamesStr = "no games"
	} else {
		gamesStr = fmt.Sprintf ("%d games", played)
	}
	ans = fmt.Sprintf ("%s: rating %d \n (%s)", nick, player.Rating, gamesStr)
	/*if played != 0 && played != len(unfinished) {
		nicks := l.GetPlayers()
		sort.Slice (nicks, func (i, j int) bool{
			return l.GetRating(nicks[i]) > l.GetRating(nicks[j])
		})
		for i, v := range nicks {
			if v == nick {
				ans += fmt.Sprintf ("\nPlace in the rating: %d", i + 1)
				break
			}
		}
	}*/
	ans += fmt.Sprintf ("\nWon %d, lost %d, unfinished %d", won, lost,  len(unfinished))
	if len(unfinished) > 0 {
		ans += fmt.Sprintf ("\nUnfinished game ids: %s", strings.Join (unfinished, ", "))
	}
	if len(contested) > 0 {
		ans += fmt.Sprintf ("\nContested games with this player (ids): %s", strings.Join (contested, ", "))
	}
	if l.GetPlayer(nick).Banned {
		ans += fmt.Sprintf ("\n%s banned", nick)
	}
	return ans
}


var (
	contested = NewBotCommand ("contested", 0, true, func (l Ladder, s *server.Server, sender string, _ []string) {
		contestedGames := l.GetGamesContested ()
		contestedIds := []string{}
		for _, game := range contestedGames {
			contestedIds = append (contestedIds, fmt.Sprintf("%d", game.Id))
		}
		s.Whisper (sender, "Contested games: "+strings.Join(contestedIds, ", "))
	})
	unfinished = NewBotCommand ("unfinished", 0, true, func (l Ladder, s *server.Server, sender string, _ []string) {
		unfinishedGames := l.GetGamesUnfinished()
		unfinishedIds := []string{}
		for _, game := range unfinishedGames {
			unfinishedIds = append (unfinishedIds, fmt.Sprintf("%d", game.Id))
		}
		s.Whisper (sender, "Unfinished games: "+strings.Join(unfinishedIds, ", "))
		s.Whisper (sender, fmt.Sprintf("Total number: %d", len (unfinishedIds)))
	})
	cancel = NewBotCommand ("cancel", 1, true, func (l Ladder, s *server.Server, sender string, a []string) {
		gameId := types.ParseInt(a[0], -1)
		game := l.GetGame (gameId)
		if game != nil {
			l.CancelGame(*game)
			fmt.Println(sender,"cancels",game)
			s.Whisper(sender, "Game canceled")
		}
	})
	cancel_old = NewBotCommand ("cancel_old", 0, true, func (l Ladder, s *server.Server, sender string, _ []string) {
		unfinishedGames := l.GetGamesUnfinishedOld(time.Now())
		for _, g := range unfinishedGames {
			l.CancelGame(g)
		}
		s.Whisper (sender, fmt.Sprintf("Canceled %d games", len (unfinishedGames)))
	})
	undo_report = NewBotCommand ("undo_report", 1, true, func (l Ladder, s *server.Server, sender string, a []string) {
		gameId := types.ParseInt(a[0], -1)
		game := l.GetGame (gameId)
		if game != nil {
			if game.TeamWon != 0 {
				l.UndoRatings(game)
				s.Whisper(sender, "Game unreported")
			} else {
				s.Whisper (sender, "Game isn't reported, nothing to undo")
			}
		} else {
			s.Whisper(sender, "No such game")
		}
	})
	force_report = NewBotCommand ("force_report", 2, true, func (l Ladder, s *server.Server, sender string, a []string) {
		gameId := types.ParseInt(a[0], -1)
		winner := a[1]
		game := l.GetGame (gameId)
		if game != nil {
			if game.Canceled {
				s.Whisper(sender, "Game canceled, you can't report it")
				return
			}
			if game.HasPlayed (winner) {
				if game.TeamWon != 0 {
					//undo ratings
					l.UndoRatings(game)
				}
				l.ReportGame (*game, winner, "admin " + sender)
				s.Whisper(sender, "Game force-reported")
			} else {
				s.Whisper(sender, fmt.Sprintf("%s hasn't played this game", winner))
			}
		} else {
			s.Whisper(sender, "Wrong id of the game")
		}
	})
	ban = NewBotCommand ("ban", 1, true, func (l Ladder, s *server.Server, sender string, command []string) {
		if l.IsSuperAdmin (sender) {
			l.Ban(strings.ToLower (command[0]))
			s.Whisper(sender, "Banned")	
		}
	})
	unban = NewBotCommand ("unban", 1, true, func (l Ladder, s *server.Server, sender string, command []string) {
		l.Unban(strings.ToLower (command[0]))
		s.Whisper(sender, "Unbanned")	
	})
	recalc = NewBotCommand ("recalc", 1, true, func (l Ladder, s *server.Server, sender string, command []string) {
		if l.IsSuperAdmin (sender) {
			l.Recalculate(strings.ToLower (command[0]))
			s.Whisper(sender, "Recalculated")
		}
	})
	uncontest = NewBotCommand ("uncontest", 1, true, func (l Ladder, s *server.Server, sender string, command []string) {
		gameId := types.ParseInt(command[0], -1)
		game := l.GetGame (gameId)
		if game != nil {
			if game.Contested {
				l.UncontestGame (*game)
				s.Whisper (sender, "Game uncontested")
			} else {
				s.Whisper(sender, "Game isn't contested")
			}
		} else {
			s.Whisper(sender, "Wrong id of the game")
		}
	})								
	stop = NewBotCommand ("stop", 0, true, func (l Ladder, s *server.Server, sender string, _ []string) {
		if l.IsSuperAdmin (sender) {
			s.Whisper(sender, "Logging out...")
			s.LeaveGame()
			panic ("Force quit from super admin "+ sender)
		}
	})
	admin_help = NewBotCommand ("admin_help", 0, true, func (_ Ladder, s *server.Server, sender string, _ []string) {
		s.Whisper(sender, "Command list:\nadmins, contested, ban, unban, dump, load, cancel, undo_report, stop, unfinished, cancel_old")
	})
	
	// for rabotyagi  //////////////////////////////////////////////////////////////////////////////
	won = NewBotCommand ("won", 1, false, func (l Ladder, s *server.Server, sender string, command []string) {
		sendPlayer := l.GetPlayer (sender)
		if sendPlayer.Banned {
			s.Whisper(sender, "You are banned, contact admins")
			return
		}
		gameId := types.ParseInt(command[0], -1)
		game := l.GetGame (gameId)
		if game != nil {
			if game.Canceled {
				s.Whisper(sender, "Game canceled, you can't report it")
				return
			}
			curTime := time.Now().Unix()
			if curTime - game.StartTime > 60 * 60 * 24 {
				s.Whisper(sender, "Game is too old (more than 24h ago), you can't report it, contact admins")
				fmt.Println(sender,"tried to report old game",gameId)
				return
			}
			if game.HasPlayed (sender) {
				if game.TeamWon == 0 {
					l.ReportGame (*game, sender, sender)
					s.Whisper(sender, "Game reported")
				} else {
					if l.HasWon (*game, sender) {
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
	})
	win = NewBotCommand ("win", 1, false, won.Comm)
	gameinfo = NewBotCommand ("game", 1, false, func (l Ladder, s *server.Server, sender string, command []string) {
		gameId := types.ParseInt(command[0], -1)
		game := l.GetGame (gameId)
		if game != nil {
			//s.Whisper (sender, "Players: "+ strings.Join (game.Players, ", ")+"\nStarted: "+time.Unix (game.StartTime,0).UTC().Format(time.UnixDate))
			if len(game.Players) != len (game.Ratings) {
				s.Whisper(sender, "Corrupted game, report to the admins!")
				return
			}
			s.Whisper(sender, "Started: "+time.Unix (game.StartTime,0).UTC().Format(time.UnixDate))
			var plWithRat strings.Builder
			for i := range game.Players {
				plWithRat.WriteString(fmt.Sprintf("%s (%d)", game.Players[i], game.Ratings[i]))
				if i != len (game.Players) - 1 {
					plWithRat.WriteString(", ")
				}
			}
			s.Whisper (sender, plWithRat.String())
			if game.Canceled {
				s.Whisper(sender, "Game canceled")
			} else {
				if game.TeamWon == 0 {
					s.Whisper (sender, "Still being played or unreported")
				} else {
					if len (game.RatAdjusts) == len(game.Players) {
						ratStrs := []string{}
						for _, rat := range game.RatAdjusts {
							ratStrs = append (ratStrs, fmt.Sprintf("%d", rat))
						}
						s.Whisper (sender, "Rating adjustments: ["+strings.Join (ratStrs, ", ") + "]")
					}
					winners := []string{}
					for _, p := range game.Players {
						if l.HasWon (*game, p) {
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
	})
	contest = NewBotCommand ("contest", 1, false, func (l Ladder, s *server.Server, sender string, command []string) {
		gameId := types.ParseInt(command[0], -1)
		game := l.GetGame (gameId)
		if game != nil {
			if game.Canceled {
				s.Whisper(sender, "Game canceled, you can't contest it")
				return
			}
			if game.HasPlayed (sender) {
				if game.Contested {
					s.Whisper (sender, "Game has already been contested")
				} else {
					l.ContestGame (*game ,sender)
					s.Whisper (sender, "Game contested")
				}
			} else {
				s.Whisper(sender, "You haven't played this game")
			}
		} else {
			s.Whisper(sender, "Wrong id of the game")
		}
	})
	playerinfo = NewBotCommand ("info", 1, false, func (l Ladder, s *server.Server, sender string, command []string) {
		s.Whisper(sender, getInfoAboutPlayer (l, command[0]))
	})
	me = NewBotCommand ("me", 0, false, func (l Ladder, s *server.Server, sender string, _ []string) {
		s.Whisper(sender, getInfoAboutPlayer (l, sender))
	})
	leaderboard = NewBotCommand ("leaderboard", 0, false, func (l Ladder, s *server.Server, sender string, _ []string) {
		lboard := l.GetTopPlayers ()
		var b strings.Builder
		b.WriteString ("\n")
		for i:=0; i < len (lboard); i = i + 1 {
			b.WriteString (fmt.Sprintf ("%d. %s: %d\n", i+1, lboard[i].Name, lboard[i].Rating))
		}
		s.Whisper (sender, b.String())
	})
	
	defaultComms = []BotCommand {contested, unfinished, cancel, undo_report, force_report, ban, unban, uncontest, stop, cancel_old, admin_help, won, win, gameinfo, contest, playerinfo, me, leaderboard, recalc}
)

func AddDefaultsToBot (b *Bot) {
	for _, v := range defaultComms {
		b.AddCommand (v)
	}
}

func AddStop (b *Bot) {
	b.AddCommand (stop)

}
