package newwesbot

import (
	"strings"
	"time"
	"math"
	"github.com/go-pg/pg/v10"
)

type Ladder interface {
	NextGameId () int
	HasWon (game Game, pl string) bool
	HasLost (game Game, pl string) bool
	ReportGame (game Game, winner, reporter string)
	RegisterGame (pl []string) int //game id
	ContestGame (game Game, contester string)
	UncontestGame (game Game)
	GetTopPlayers () []Player
	GetPlayer (pl string) Player
	SavePlayer (p Player)
	IsAdmin (pl string) bool
	IsSuperAdmin (pl string) bool
	Ban (pl string)
	Unban (pl string)
	CancelGame (game Game)
	UndoRatings (game *Game)
	GetGame (id int) *Game
	GetGamesContested () []Game
	GetGamesOf (pl string) []Game
	Recalculate (pl string) 
	GetGamesUnfinished () []Game
	GetGamesUnfinishedOld (at time.Time) []Game
	IsQualified (pl string) bool
}
/*
*/
type GenericLadder struct {
	db *pg.DB
	Admins []string
	p LadderParameters
	strictMode bool
}

var (
	_ Ladder = GenericLadder{}
)

func EloRating (rats []int, winner int) []int {
	points := []float64{0.0, 0.0}
	points[winner] = 1.0
	expected := []float64 {0.0, 0.0}
	expected[0] = 1.0 / (1.0 + math.Pow (10.0, float64(rats[1] - rats[0])/400.0))
	expected[1] = 1.0 / (1.0 + math.Pow (10.0, float64(rats[0] - rats[1])/400.0))
	adjs := []int {0, 0}
	for i, _ := range expected {
		k := 30.0
		if rats[i] > 1800 {
			k = 10.0
		} else if rats[i] > 1650{
			k = 15.0
		} else if rats[i] > 1400 {
			k = 20.0
		}
		adjs[i] = int (k * (points[i] - expected[i]))
		if adjs[i] == 0 {
			if i == winner {
				adjs[i] = 1//at least one point
			} else {
				adjs[i] = -1
			}
		}
	}
	return adjs
}

func NewGenericLadder(db_ *pg.DB, admins []string, pp LadderParameters, strictMode bool) Ladder{
	return GenericLadder {
		db: db_,
		Admins: admins,
		p: pp,
		strictMode: strictMode,
	}
}

func (l GenericLadder) NextGameId () int {
	game := new (Game)
	err := l.db.Model (game).Last()
	if err != nil {
		return 1
	}
	return game.Id + 1
}

func (l GenericLadder) RegisterGame (pl []string) int {
	game := new (Game)
	game.StartTime = time.Now().Unix()
	nicks := make ([]string, len (pl))
	ratings := make([]int, len(pl))
	for i:=0;i<len(pl);i++ {
		nicks[i] = strings.ToLower (pl[i])
		player := l.GetPlayer (pl[i])
		ratings[i] = player.Rating
	}
	game.Players = nicks
	game.Ratings = ratings
	_, err := l.db.Model (game).Insert()
	if err != nil {
		panic (err)
	}
	return game.Id
}

func (l GenericLadder) ReportGame (game Game, winner, reporter string) {
	if game.TeamWon != 0 {
		panic ("Reporting reported game")
	}
	teamWon, teamLost := l.p.TeamFun (game.Players, winner)
	if teamWon == -1 || teamLost == -1 {
		panic ("winner didn't played the game")
	}
	game.TeamWon = teamWon
	pls := []Player{}
	for _, plnick := range game.Players {
		pls = append(pls, l.GetPlayer (plnick))
	}
	ratDiff := l.p.RatingFun (game.Ratings, teamWon)
	for i := range pls {
		pls[i].Rating = pls[i].Rating + ratDiff[i]
	}
	
	game.RatAdjusts = ratDiff
	game.Reporter = reporter
	tx, err := l.db.Begin()
	defer tx.Close()
	if err != nil {
		panic (err)
	}
	_, err = tx.Model(&game).WherePK().Update()
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	for _, pl := range pls {
		_, err = tx.Model(&pl).OnConflict("(name) DO UPDATE").Insert()
		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}
	tx.Commit()
}

func (l GenericLadder) Ban (pl string) {
	player := l.GetPlayer (pl)
	player.Banned = true
	l.SavePlayer (player)
}

func (l GenericLadder) Unban (pl string) {
	player := l.GetPlayer (pl)
	player.Banned = false
	l.SavePlayer (player)
}

func (l GenericLadder) GetPlayer (pl string) Player {
	pl = strings.ToLower (pl)
	player := &Player {
		Name: pl,
	}
	err := l.db.Model (player).WherePK().Select()
	if err != nil {
		return Player {
			Name: pl,
			Rating: 1500,
		}
	}
	return *player
}

func (l GenericLadder) SavePlayer (pl Player) {
	_, err := l.db.Model (&pl).OnConflict("(name) DO UPDATE").Insert()
	if err != nil {
		panic (err)
	}
}

func (l GenericLadder) Recalculate (pl string) {
	pl = strings.ToLower (pl)
	player := l.GetPlayer (pl)
	
	games := l.GetGamesOf (pl)
	tx, err := l.db.Begin()
	defer tx.Close()
	if err != nil {
		panic ("Faileddd")
	}
	rereportGame := func (game Game, teamWon int, reporter string, curRat int) int {
		game.TeamWon = teamWon
		pls := []Player{}
		for _, plnick := range game.Players {
			pls = append(pls, l.GetPlayer (plnick))
		}
		for i := range game.Ratings {
			if game.Players[i] == pl {
				game.Ratings[i] = curRat
				pls[i].Rating = curRat
				break
			}
		}
		ratDiff := l.p.RatingFun (game.Ratings, teamWon)
		for i := range pls {
			pls[i].Rating = pls[i].Rating + ratDiff[i]
		}
		
		game.RatAdjusts = ratDiff
		game.Reporter = reporter
		_, err := tx.Model(&game).WherePK().Update()
		if err != nil {
			tx.Rollback()
			panic(err)
		}
		for i := range pls {
			if pls[i].Name == pl {
				return pls[i].Rating
			}
		}
		panic ("Failed")
	}
	rat := 1500
	for i := range games {
		team := games[i].TeamWon
		repo := games[i].Reporter
		if team != 0 {
			rat = rereportGame (games[i], team, repo+" (recalced)", rat)
		}
	}
	tx.Commit()
	player.Rating = rat
	l.SavePlayer (player)
}

func (l GenericLadder) UndoRatings (game *Game) {
	if game == nil || game.TeamWon == 0 || len(game.Players) != len(game.RatAdjusts){
		return
	}
	adjusts := game.RatAdjusts
	game.RatAdjusts = []int{}
	game.TeamWon = 0
	game.Reporter = ""
	tx, err := l.db.Begin()
	defer tx.Close()
	_, err = tx.Model(game).WherePK().Update()
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	for i, n := range game.Players {
		pl := l.GetPlayer (n)
		pl.Rating = pl.Rating - adjusts[i]
		_, err = tx.Model(&pl).OnConflict("(name) DO UPDATE").Insert()
		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}
	tx.Commit()
}

func (l GenericLadder) CancelGame (game Game) {
	l.UndoRatings (&game)
	game.Canceled = true
	_, err := l.db.Model (&game).WherePK().Update()
	if err != nil {
		panic (err)
	}
}

func (l GenericLadder) ContestGame (game Game, contester string) {
	game.Contested = true
	game.ContestedBy = contester
	_, err := l.db.Model (&game).WherePK().Update()
	if err != nil {
		panic (err)
	}
}

func (l GenericLadder) UncontestGame (game Game) {
	game.Contested = false
	_, err := l.db.Model (&game).WherePK().Update()
	if err != nil {
		panic (err)
	}
}

func (l GenericLadder) GetGame (id int) *Game {
	game := &Game {
		Id: id,
	}
	err := l.db.Model (game).WherePK().Select()
	if err != nil {
		return nil
	}
	return game
}

func (l GenericLadder) GetGamesUnfinished () []Game {
	var games []Game
	err := l.db.Model (&games).Where("Team_Won = ?", 0).Where("Canceled = ?", false).Order ("id ASC").Select()
	if err != nil {
		panic(err)
	}
	return games
}

func (l GenericLadder) GetGamesOf (pl string) []Game {
	var games []Game
	err := l.db.Model (&games).Where("? = Any(Players)", strings.ToLower(pl)).Where("Canceled = ?", false).Order ("id ASC").Select()
	if err != nil {
		panic(err)
	}
	return games
}

func (l GenericLadder) GetGamesContested () []Game {
	var games []Game
	err := l.db.Model (&games).Where("Contested = ?", true).Where("Canceled = ?", false).Order ("id ASC").Select()
	if err != nil {
		panic(err)
	}
	return games
}

func (l GenericLadder) GetGamesUnfinishedOld (at time.Time) []Game {
	unix := at.Unix() - 60 * 60 * 24
	var games []Game
	err := l.db.Model (&games).Where("Team_Won = ?", 0).Where("Start_Time < ?",unix).Where("Canceled = ?", false).Select()
	if err != nil {
		panic(err)
	}
	return games
}

func (l GenericLadder) GetTopPlayers () []Player {
	var players []Player
	err := l.db.Model (&players).Where("Banned = ?", false).Order ("rating DESC").Limit(10).Select()
	if err != nil {
		panic (err)
	}
	return players
}

func (l GenericLadder) HasLost (game Game, pl string) bool{
	team := game.TeamWon
	if team == 0 {
		return false
	}
	return l.p.HasLost (game.Players, team, pl)
}

func (l GenericLadder) HasWon (game Game, pl string) bool{
	team := game.TeamWon
	if team == 0 {
		return false
	}
	return l.p.HasWon (game.Players, team, pl)
}

func (l GenericLadder) IsSuperAdmin (pl string) bool {
	return l.Admins[0] == strings.ToLower (pl)
}

func (l GenericLadder) IsAdmin (pl string) bool{
	for _, adm := range l.Admins {
		if strings.ToLower (pl) == adm {
			return true
		}
	}
	return false
}

func (l GenericLadder) IsQualified (pl string) bool {
	if !l.strictMode {
		return true
	}
	var unreportedGames []Game
	err := l.db.Model (&unreportedGames).Where("? = Any(Players)", strings.ToLower(pl)).Where("Canceled = ?", false).Where("Team_Won = ?", 0).Order ("id ASC").Select()
	if err != nil {
		panic(err)
	}
	counter := 0
	for _, g := range unreportedGames {
		if g.ContestedBy == "" {
			counter += 1
		}
		if counter >= 5 {
			return false
		}
	}
	return true
}
