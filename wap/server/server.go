// This file is part of Fastbot.
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

package server

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"wap/server/hash"
	"math/rand"
	"net"
	"strconv"
	"time"

	serverTypes "wap/server/types"
	"wap/types"
	e "go-wesnoth/era"
	"go-wesnoth/mod"
	"go-wesnoth/addon"
	"go-wesnoth/game"
	"go-wesnoth/scenario"
	"go-wml"
)

type Server struct {
	hostname      string
	port          uint16
	version       string
	Username      string
	password      string
	game          []byte
	Observers     types.StringList
	timeout       time.Duration
	err           error
	conn          net.Conn
	disconnecting bool
	Sides         serverTypes.SideList
	TimerEnabled  bool
	InitTime      int
	TurnBonus     int
	ReservoirTime int
	ActionBonus   int
	ForceFinish bool
	InGame bool
}

func NewServer(hostname string, port uint16, version string, username string,
	password string, 
	timerEnabled bool,
	initTime int, turnBonus int, reservoirTime int, actionBonus int,
	timeout time.Duration, forceFinish bool) *Server {
	s := Server{
		hostname:      hostname,
		port:          port,
		version:       version,
		Username:      username,
		password:      password,
		timeout:       timeout,
		TimerEnabled:  timerEnabled,
		InitTime:      initTime,
		TurnBonus:     turnBonus,
		ReservoirTime: reservoirTime,
		ActionBonus:   actionBonus,
		ForceFinish: forceFinish,
		InGame: false,
	}
	s.Sides = serverTypes.SideList{}
	return &s
}

func (s *Server) Connect () error {
	return s.ConnectEnhanced(false)
}

func (s *Server) ConnectEnhanced(tlsFlag bool) error {
	// Set up a TCP connection
	s.conn, s.err = net.Dial("tcp", s.hostname+":"+strconv.Itoa(int(s.port)))
	if s.err != nil {
		return s.err
	}
	//s.conn.SetDeadline(time.Now().Add(s.timeout))
	// Init the connection to the server
	if !tlsFlag {
		s.conn.Write([]byte{0, 0, 0, 0})
	} else {
		s.conn.Write([]byte{0, 0, 0, 1})
	}
	var buffer []byte
	if buffer, s.err = s.read(4); s.err != nil {
		return s.err
	}
	fmt.Println("buffer_info", binary.BigEndian.Uint32(buffer))
	if tlsFlag {//TODO: switch depending on server answer, not alwat=ys
		s.conn = tls.Client (s.conn, &tls.Config{
			//MinVersion: 0x0304,
			ServerName: "server.wesnoth.org", //TODO: fix
		})
	}
	// Expects the server to ask for a version, otherwise return an error
	if data, _ := s.receiveData(); bytes.Equal(data, wml.EmptyTag("version").Bytes()) {
		s.sendData((&wml.Tag{"version", wml.NewDataAttrs(wml.AttrMap{"version": s.version})}).Bytes())
	} else {
		return errors.New("Expects the server to request a version, but it doesn't.")
	}
	// Expects the server to require the log in step, otherwise return an error
	{
		rawData, _ := s.receiveData()
		fmt.Println(string(rawData))
		data := wml.ParseData(rawData)
		switch {
		case bytes.Equal(rawData, wml.EmptyTag("mustlogin").Bytes()):
			s.sendData((&wml.Tag{"login", wml.NewDataAttrs(wml.AttrMap{"selective_ping": "1", "username": s.Username})}).Bytes())
		case data.ContainsTag("redirect"):
			redirectTags := data.GetTags ("redirect")
			if len(redirectTags) == 1 {
				redirect := redirectTags[0]
				host := redirect.GetAttr("host")
				port := redirect.GetAttr("port")
				if host != "" && port != "" {
					portInt, err := strconv.Atoi(port)
					if err == nil {
						s.hostname = host
						s.port = uint16(portInt)
						fmt.Println("REDIRECT " + host + " " + port)
						return s.ConnectEnhanced(tlsFlag)
					}
				}
			}
			fallthrough
		default:
			return errors.New("Expects the server to require a log in step, but it doesn't.")
		}
	}
	rawData, _ := s.receiveData()
	data := wml.ParseData(rawData)
	fmt.Println(data)
	switch {
	case data.ContainsTag("error"):
		fmt.Println("ERROR CASE")
		
		if errorTag, ok := data.GetTag("error"); ok {
			fmt.Println (errorTag.GetAttr("error_code"))
			code := errorTag.GetAttr("error_code")
			switch code {
				case "200", "201":// 201 = already logged in before, kick that session out
					if errorTag.GetAttr("password_request") == "yes"{
						salt := errorTag.GetAttr("salt")
						var passwordHash string
						if errorTag.GetAttr("phpbb_encryption") == "yes" {
							if hash.MD5Prefix (salt) {
								passwordHash = hash.Sum(s.password, salt)
							} else {
								pass, errb := hash.WesnothBcrypt (salt, s.password)
								if errb != nil {
									return errb
								}
								passwordHash = pass
							}
						} else if tlsFlag {
							passwordHash = s.password
						} else {
							panic ("No hash but no tls")
						}
						qqq := (&wml.Tag{Name: "login", Data: wml.NewDataAttrs(wml.AttrMap{"username": s.Username, "password": passwordHash})}).Bytes()
						//fmt.Println("La nina", string(qqq))
						s.sendData(qqq)
						fmt.Println("Yo")
						rawData, s.err = s.receiveData()
						data = wml.ParseData(rawData)
						fmt.Println("newdata", data)
						goto nextCase
					}
				case "105":
					if errorTag.Contains ("message") {
						return errors.New(errorTag.GetAttr ("message"))
					} else {
						return errors.New("The nickname is not registered. This server disallows unregistered nicknames.")
					}
			}
		}
				
		break
	nextCase:
		
		fallthrough
	case data.ContainsTag("join_lobby") || data.ContainsTag("mustlogin"):
		return nil
	default:
		return errors.New("An unknown error occurred")
	}
	return nil
}

func (s *Server) HostGame(sc scenario.Scenario, era e.Era, mods []mod.Mod, addons []addon.Addon, name, pass string) {
	if s.InGame {
		return
	}
	g := game.NewGame(name,
		sc,
		era, mods, addons,
		s.TimerEnabled, s.InitTime, s.TurnBonus, s.ReservoirTime, s.ActionBonus,
		s.version)
	s.HostGameFromTemplate (sc, g, name, pass)
}

func (s *Server) HostGameFromTemplate(sc scenario.Scenario, g game.Game, name, pass string) {
	if s.InGame {
		return
	}

	s.Sides = serverTypes.SideList{}
	for i, v := range sc.Sides {
		side := &serverTypes.Side{Side: i + 1, Color: colors[i%len(colors)], Controller: v}
		if v == "ai" {
			side.Player = s.Username
		}
		s.Sides = append (s.Sides, side)
		fmt.Println (side, i, v)
	}
	s.Observers = types.StringList{}
	g.Title = name
	s.game = g.Bytes()
	s.InGame = true
	s.sendData(
		(&wml.Tag{"create_game", wml.NewDataAttrs(wml.AttrMap{"name": name,
		"password": pass})}).Bytes())
	s.sendData(s.game)
}

func defaultShuffler (players []string) {
	rand.Shuffle(len (players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
}

type Shuffler func ([]string)

func (s *Server) StartGame (era e.Era, units map[string]*wml.Data, shuffle bool) ([]string, error) {
	return s.StartGameEx (era, units, shuffle, defaultShuffler)
}

func (s *Server) StartGameEx(era e.Era, units map[string]*wml.Data, shuffle bool, shuffler Shuffler) ([]string, error){
	err := s.sendData(wml.EmptyTag("stop_updates").Bytes())
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UTC().UnixNano())
	qq := []int {}
	for i := range era.Factions {
		qq = append (qq, i)
	}
	rand.Shuffle(len (era.Factions), func(i, j int) {
		qq[i], qq[j] = qq[j], qq[i]
	})
	//TODO: doesn't work with  ai sides
	if shuffle {
		if shuffler == nil {
			shuffler = defaultShuffler
		}
		players := []string{}
		for _, side := range s.Sides {
			if side.Controller == "human" {
				players = append (players, side.Player)
			}
		}
		fmt.Println (players)
		shuffler(players)
		cnt := 0
		for _, side := range s.Sides {
			if side.Controller == "human" {
				side.Player = players[cnt]
				cnt++
			}
		}
		fmt.Println (players)
	}
	/*var sidesData = wml.Multiple{}
	for i, side := range s.Sides {
		sidesData = append (sidesData, wml.Data{"index": i, "side": insertFaction(side, era.Factions[qq[i%len(qq)]], units)})
	}
	data := wml.Data{"scenario_diff": wml.Data{"change_child": wml.Data{
		"index": 0,
		"scenario": wml.Data{"change_child": sidesData},
	},
	}}*/
	/*
	sidesData := wml.NewData()
	for i, side := range s.Sides {
		tmp := &wml.Data{Attrs: wml.AttrMap{"index": i}, Tags: []*wml.Tag{&wml.Tag{"side", insertFaction(side, era.Factions[qq[i%len(qq)]], units)}}}
		sidesData.AddTagByName ("change_child", tmp)
	}
	scenarioTag := &wml.Tag{"scenario", sidesData}
	changeChildTag := &wml.Tag{"change_child", &wml.Data {Attrs: wml.AttrMap {"index": 0}, Tags: []*wml.Tag{scenarioTag}}}*/
	sidesData := wml.NewData()
	factionIDs := []string{}
	for i, side := range s.Sides {
		tmp := wml.NewDataAttrsTags(wml.AttrMap{"index": i}, &wml.Tag{"side", insertFaction(side, era.Factions[qq[i%len(qq)]], units)})
		sidesData.AddTagByName ("change_child", tmp)
		factionIDs = append (factionIDs, getFactionID(era.Factions[qq[i%len(qq)]]))
	}
	scenarioTag := &wml.Tag{"scenario", sidesData}
	changeChildTag := &wml.Tag{"change_child", wml.NewDataAttrsTags(wml.AttrMap {"index": 0},scenarioTag)}
	data := wml.NewDataTags(&wml.Tag{"scenario_diff", wml.NewDataTags (changeChildTag)}) 
	
	err = s.sendData(data.Bytes())
	if err != nil{
		return nil, err
	}
	return factionIDs, s.sendData(wml.EmptyTag("start_game").Bytes())
}

func (s *Server) LeaveGame() {
	if s.InGame {
		s.sendData(wml.EmptyTag("leave_game").Bytes())
	}
	s.InGame = false
}

func (s *Server) SetSidePlayer (side int, name string, ready bool) bool {
	if s.Sides.HasSide(side) && !s.Sides.HasPlayer(name) {
		s.ChangeSide(side, "insert", wml.NewDataAttrs(wml.AttrMap{"current_player": name, "name": name, "player_id": name}))
		s.Sides.Side(side).Player = name
		s.Sides.Side(side).Ready = ready
		return true
	}
	return false
}

func (s *Server) ClearSide (side int) {
	s.ChangeSide(side, "delete", wml.NewDataAttrs(wml.AttrMap{"current_player": "x", "name": "x", "player_id": "x"}))
	Sidestruct := s.Sides.Side(side)
	Sidestruct.Ready = false
	Sidestruct.Player = ""
}

func (s *Server) Disconnect() {
	time.Sleep(time.Second * 3)
	s.disconnecting = true
	s.InGame = false
}

func (s *Server) GetServerInput(maxSize int) (*wml.Data, error) {
	data, err := s.receiveData()
	if err != nil {
		return &wml.Data{}, err
	}
	if maxSize > 0 && len(data) > maxSize {
		//fmt.Println ("Server sent bytes:", len(data), ", rejecting")
		return wml.NewData(), nil
	}
	//fmt.Println ("Server sent bytes:", len(data), ", accepting")
	return wml.ParseData(data), nil
}

func (s *Server) ChangeSide(side int, command string, data *wml.Data) error{
	tagToSend := wml.NewDataTags (&wml.Tag{"scenario_diff", wml.NewDataTags(
	  &wml.Tag{"change_child", wml.NewDataAttrsTags (wml.AttrMap{"index": 0}, 
	    &wml.Tag{"scenario", wml.NewDataTags(
	      &wml.Tag{"change_child",wml.NewDataAttrsTags(wml.AttrMap{"index": side - 1},
	        &wml.Tag{"side", wml.NewDataTags(&wml.Tag{command, data})})})})})})
		                                  
	return s.sendData(tagToSend.Bytes())
}

func (s *Server) Message(text string) error{
	for _, v := range SplitMessage(wml.EscapeString(text)) {
		err := s.sendData(wml.NewDataTags(&wml.Tag{"message", wml.NewDataAttrs(wml.AttrMap{"message": v, "room": "this game", "sender": s.Username})}).Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) InGameMessage(text string) error{
	for _, v := range SplitMessage(wml.EscapeString(text)) {
		err := s.sendData(wml.NewDataTags(&wml.Tag{"turn", wml.NewDataTags(&wml.Tag{"command", wml.NewDataAttrsTags(wml.AttrMap{"undo": "no"}, &wml.Tag{"speak", wml.NewDataAttrs(wml.AttrMap{"message": v, "id": s.Username, "time": fmt.Sprintf("%d", time.Now().Unix())})})})}).Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Whisper(receiver string, text string) error{
	for _, v := range SplitMessage(wml.EscapeString(text)) {
		err := s.sendData(wml.NewDataTags(&wml.Tag{"whisper", wml.NewDataAttrs(wml.AttrMap{"sender": s.Username, "receiver": receiver, "message": v})}).Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}


func (s *Server) Error() error {
	return s.err
}

func (s *Server) receiveData() ([]byte, error) {
	buffer, err := s.read(4)
	if err != nil {
		return nil, err
	}
	if len(buffer) < 4 {
		return nil, nil
	}
	size := int(binary.BigEndian.Uint32(buffer))
	forBuf, _ := s.read(size)
	reader, _ := gzip.NewReader(bytes.NewBuffer(forBuf))
	var result []byte
	if result, s.err = ioutil.ReadAll(reader); s.err != nil {
		return nil, s.err
	}
	if s.err = reader.Close(); s.err != nil {
		return nil, s.err
	}
	return result, nil
}

func (s *Server) sendData(data []byte) error {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	gz.Write([]byte(data))
	gz.Close()

	var length int = len(b.Bytes())
	_, s.err = s.conn.Write([]byte{byte((length / (256*256*256))%256), byte((length / (256*256))%256), byte((length / 256)%256), byte(length % 256)})
	if s.err != nil {
		return s.err
	}
	_, s.err = s.conn.Write(b.Bytes())
	return s.err
}

func (s *Server) read(n int) ([]byte, error)  {
	result := []byte{}
	count := 0
	for count < n {
		buffer := make([]byte, n-count)
		var num int
		s.conn.SetReadDeadline(time.Now().Add(time.Minute * 5))
		num, s.err = s.conn.Read(buffer)
		if s.err != nil {
			return nil, s.err
		}
		count += num
		result = append(result, buffer[:num]...)
	}
	return result, nil
}
