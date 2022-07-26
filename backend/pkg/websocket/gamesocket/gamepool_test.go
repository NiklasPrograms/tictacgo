package gamesocket

import (
	"testing"

	"github.com/NiklasPrograms/tictacgo/backend/pkg/game"
	"github.com/NiklasPrograms/tictacgo/backend/pkg/websocket"
	"github.com/NiklasPrograms/tictacgo/backend/pkg/websocket/testutils"
)

func setupTest(t *testing.T) (func(t *testing.T), *GamePool) {
	t.Log("Setting up testing")

	pool := NewGamePool(NewSequentialChannelStrategy())

	return func(t *testing.T) {
		t.Log("Tearing down testing")
	}, pool
}

func createTestClient(pool *GamePool) *GameClient {
	client := &GameClient{
		Pool: pool,
		conn: testutils.NewConnMock(),
	}

	pool.Register(client)

	return client
}

func startGame(pool *GamePool) (websocket.Client, websocket.Client) {
	clientX, clientO := createTestClient(pool), createTestClient(pool)

	messageX := GameMessage{SELECT_CHARACTER, game.X.String(), clientX}
	messageO := GameMessage{SELECT_CHARACTER, game.O.String(), clientO}

	pool.Broadcast(messageX)
	pool.Broadcast(messageO)

	message := GameMessage{START_GAME, 0, clientX}
	pool.Broadcast(message)

	return clientX, clientO
}

func chooseSquare(pool *GamePool, c websocket.Client, position game.Position) {
	message := GameMessage{CHOOSE_SQUARE, position, c}
	pool.Broadcast(message)
}

func TestRegisterClient(t *testing.T) {
	teardown, pool := setupTest(t)
	defer teardown(t)

	clientsInPool := len(pool.Clients())
	if clientsInPool != 0 {
		t.Fatalf("Expected no clients in pool, got %d", clientsInPool)
	}

	createTestClient(pool)
	clientsInPool = len(pool.Clients())
	if clientsInPool != 1 {
		t.Fatalf("Expected 1 client in pool, got %d", clientsInPool)
	}

	createTestClient(pool)
	clientsInPool = len(pool.Clients())
	if clientsInPool != 2 {
		t.Log(pool.Clients())
		t.Fatalf("Expected 2 clients in pool, got %d", clientsInPool)
	}
}

func TestUnregisterClient(t *testing.T) {
	teardown, pool := setupTest(t)
	defer teardown(t)

	createTestClient(pool)
	createTestClient(pool)

	clientToUnregister := createTestClient(pool)
	pool.Unregister(clientToUnregister)

	clientsInPool := len(pool.Clients())
	if clientsInPool != 2 {
		t.Fatalf("Expected 2 clients in pool, got %d", clientsInPool)
	}
}

func TestShouldBeCharacterX(t *testing.T) {
	teardown, pool := setupTest(t)
	defer teardown(t)

	client := createTestClient(pool)

	message := GameMessage{SELECT_CHARACTER, game.X.String(), client}

	pool.Broadcast(message)

	if pool.xClient != client {
		t.Errorf("Expected client to be xClient, but got %v", pool.xClient)
	}
}

func TestShouldBeCharacterO(t *testing.T) {
	teardown, pool := setupTest(t)
	defer teardown(t)

	client := createTestClient(pool)

	message := GameMessage{SELECT_CHARACTER, game.O.String(), client}

	pool.Broadcast(message)

	if pool.oClient != client {
		t.Errorf("Expected client to be oClient, but got %v", pool.oClient)
	}
}

func TestShouldNotChangeCharacterIfCharacterIsAlreadyTaken(t *testing.T) {
	teardown, pool := setupTest(t)
	defer teardown(t)

	client1 := createTestClient(pool)
	message1 := GameMessage{SELECT_CHARACTER, game.X.String(), client1}
	pool.Broadcast(message1)

	client2 := createTestClient(pool)
	message2 := GameMessage{SELECT_CHARACTER, game.X.String(), client2}
	pool.Broadcast(message2)

	want := game.EMPTY
	got := pool.Clients()[client2]

	if want != got {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGameShouldNotStartWhenNoCharactersSelected(t *testing.T) {
	teardown, pool := setupTest(t)
	defer teardown(t)

	if pool.game.IsStarted() {
		t.Fatal("Game should not have started yet")
	}

	client := createTestClient(pool)

	message := GameMessage{START_GAME, 0, client}
	pool.Broadcast(message)

	if pool.game.IsStarted() {
		t.Errorf("Game should still not have started, despite the Start Game message, since both characters must've been selected")
	}
}

func TestGameShouldBeAbleToStartWhenBothCharactersAreSelected(t *testing.T) {
	teardown, pool := setupTest(t)
	defer teardown(t)

	client1 := createTestClient(pool)
	message1 := GameMessage{SELECT_CHARACTER, game.X.String(), client1}
	pool.Broadcast(message1)

	client2 := createTestClient(pool)
	message2 := GameMessage{SELECT_CHARACTER, game.O.String(), client2}
	pool.Broadcast(message2)

	message := GameMessage{START_GAME, 0, client1}
	pool.Broadcast(message)

	if !pool.game.IsStarted() {
		t.Errorf("Game should be started")
	}
}

func TestOCannotChooseSquareWhenItIsX(t *testing.T) {
	teardown, pool := setupTest(t)
	defer teardown(t)

	_, clientO := startGame(pool)

	originalBoard := pool.game.Board()

	chooseSquare(pool, clientO, game.CENTER)

	want := originalBoard
	got := pool.game.Board()

	if want != got {
		t.Errorf("Wanted %v, got %v", want, got)
	}
}

func TestXCannotChooseSquareWhenItIsO(t *testing.T) {
	teardown, pool := setupTest(t)
	defer teardown(t)

	clientX, _ := startGame(pool)

	chooseSquare(pool, clientX, game.CENTER)

	boardAfterFirstPlay := pool.game.Board()

	chooseSquare(pool, clientX, game.BOTTOM_CENTER)

	want := boardAfterFirstPlay
	got := pool.game.Board()

	if want != got {
		t.Errorf("Wanted %v, got %v", want, got)
	}
}

func TestSpectatorCannotStartGame(t *testing.T) {
	teardown, pool := setupTest(t)
	defer teardown(t)

	clientX := createTestClient(pool)
	clientO := createTestClient(pool)
	clientSpectator := createTestClient(pool)

	messageX := GameMessage{SELECT_CHARACTER, game.X.String(), clientX}
	messageO := GameMessage{SELECT_CHARACTER, game.O.String(), clientO}

	pool.Broadcast(messageX)
	pool.Broadcast(messageO)

	startGameMessage := GameMessage{START_GAME, 0, clientSpectator}
	pool.Broadcast(startGameMessage)

	if pool.game.IsStarted() {
		t.Errorf("Game should not have started, since it was started by the spectator")
	}
}

func TestClientCannotChooseBothCharacters(t *testing.T) {
	teardown, pool := setupTest(t)
	defer teardown(t)

	client := createTestClient(pool)

	messageX := GameMessage{SELECT_CHARACTER, game.X.String(), client}
	pool.Broadcast(messageX)

	if pool.xClient != client {
		t.Errorf("Client should have selected X")
	}

	messageO := GameMessage{SELECT_CHARACTER, game.O.String(), client}
	pool.Broadcast(messageO)

	if pool.oClient == client {
		t.Errorf("Client should not be able to select O, when they already selected X")
	}

	if pool.xClient != client {
		t.Errorf("Client should still have selected X")
	}
}
