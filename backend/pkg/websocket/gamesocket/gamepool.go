package gamesocket

import (
	"fmt"
	"net/http"

	"github.com/NiklasPrograms/tictacgo/backend/pkg/game"
	"github.com/NiklasPrograms/tictacgo/backend/pkg/websocket"
)

type GamePool struct {
	register        chan websocket.Client
	unregister      chan websocket.Client
	clients         map[websocket.Client]game.SquareCharacter
	broadcast       chan GameMessage
	game            game.GameService
	channelStrategy ChannelStrategy
}

var _ websocket.Pool = new(GamePool)

func NewGamePool(cs ChannelStrategy) *GamePool {
	return &GamePool{
		register:        make(chan websocket.Client),
		unregister:      make(chan websocket.Client),
		clients:         make(map[websocket.Client]game.SquareCharacter),
		broadcast:       make(chan GameMessage),
		game:            game.NewGame(),
		channelStrategy: cs,
	}
}

func (p *GamePool) NewClient(w http.ResponseWriter, r *http.Request) websocket.Client {
	clientName := r.URL.Query().Get("name")
	if clientName == "" {
		clientName = "Unknown"
	}

	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
	}

	client := &GameClient{
		name: clientName,
		conn: conn,
		Pool: p,
	}

	return client
}

func (p *GamePool) Broadcast(m GameMessage) {
	p.channelStrategy.broadcast(p, m)
}

func (p *GamePool) Register(c websocket.Client) {
	p.channelStrategy.register(p, c)
}

func (p *GamePool) Unregister(c websocket.Client) {
	p.channelStrategy.unregister(p, c)
}

func (p *GamePool) Clients() map[websocket.Client]game.SquareCharacter {
	return p.clients
}

func (p *GamePool) isCharacterTaken(character game.SquareCharacter) bool {
	for _, v := range p.clients {
		if v == character {
			return true
		}
	}
	return false
}

func (g *GamePool) registerCharacter(client websocket.Client, character game.SquareCharacter) error {
	if g.isCharacterTaken(character) {
		return fmt.Errorf("Character %v is already taken", character)
	}
	g.clients[client] = character
	return nil
}

func (p *GamePool) broadcastResponse(response GameResponse) error {
	for client := range p.clients {
		if err := client.Conn().WriteJSON(response); err != nil {
			return err
		}
	}
	return nil
}

func (pool *GamePool) broadcastGameIsOver() {
	gameOverResponse := GameResponse{GAME_OVER, true}

	result := pool.game.GetResult()
	resultResponse := GameResponse{RESULT, result}

	pool.broadcastResponse(gameOverResponse)
	pool.broadcastResponse(resultResponse)
}

func (pool *GamePool) respondStartGame() (GameResponse, error) {
	var response GameResponse

	board := pool.game.StartGame()

	response.Command = BOARD
	response.Body = board

	return response, nil
}

func (pool *GamePool) respondChooseSquare(message GameMessage) (GameResponse, error) {
	var response GameResponse

	position, err := game.ParsePosition(message.Content)
	if err != nil {
		return response, err
	}

	board := pool.game.ChooseSquare(position)

	response.Command = BOARD
	response.Body = board

	return response, nil
}

func (pool *GamePool) respondGetBoard() (GameResponse, error) {
	var response GameResponse

	board := pool.game.Board()

	response.Command = BOARD
	response.Body = board

	return response, nil
}

func (pool *GamePool) respondSelectCharacter(message GameMessage) (GameResponse, error) {
	var response GameResponse

	client := message.Client
	character, err := game.ParseSquareCharacter(message.Content)
	if err != nil {
		return response, err
	}

	if err := pool.registerCharacter(client, character); err != nil {
		return response, err
	}

	response.Command = CHARACTER_SELECTED
	response.Body = character

	return response, nil
}

func (pool *GamePool) executeMessage(message GameMessage) (GameResponse, error) {
	switch message.Instruction {
	case START_GAME:
		return pool.respondStartGame()
	case CHOOSE_SQUARE:
		return pool.respondChooseSquare(message)
	case GET_BOARD:
		return pool.respondGetBoard()
	case SELECT_CHARACTER:
		return pool.respondSelectCharacter(message)
	}

	return GameResponse{}, fmt.Errorf("GameInstruction could not be found: %v", message.Instruction)
}

func (pool *GamePool) respond(message GameMessage) error {
	response, err := pool.executeMessage(message)

	if err != nil {
		return err
	}

	pool.broadcastResponse(response)

	if pool.game.IsGameOver() {
		pool.broadcastGameIsOver()
	}

	return nil
}

func (pool *GamePool) Start() {
	for {
		select {
		case client := <-pool.register:
			pool.clients[client] = game.EMPTY
		case client := <-pool.unregister:
			delete(pool.clients, client)
		case message := <-pool.broadcast:
			if err := pool.respond(message); err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}
