package gamesocket

import (
	"encoding/json"

	"github.com/NiklasPrograms/tictacgo/backend/pkg/game"
)

type GameResponse struct {
	ResponseType ResponseType `json:"responseType"`
	Body         any          `json:"body"`
}

type ResponseType uint8

var _ json.Marshaler = new(ResponseType) // Command implements Marshaler interface

const (
	BOARD ResponseType = iota + 1
	GAME_OVER
	RESULT
	NEW_MESSAGE
	CHARACTER_SELECTED
	GAME_STARTED
	WELCOME
)

type WelcomeResponse struct {
	IsGameStarted bool       `json:"isGameStarted"`
	XClient       string     `json:"xClient"`
	OClient       string     `json:"oClient"`
	Board         game.Board `json:"board"`
}

var (
	ResponseTypeName = map[uint8]string{
		1: "board",
		2: "game over",
		3: "result",
		4: "new message",
		5: "character selected",
		6: "game started",
		7: "welcome",
	}
	ResponseTypeValue = map[string]uint8{
		"board":              1,
		"game over":          2,
		"result":             3,
		"new message":        4,
		"character selected": 5,
		"game started":       6,
		"welcome":            7,
	}
)

func (r ResponseType) String() string {
	return ResponseTypeName[uint8(r)]
}

func (r ResponseType) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}
