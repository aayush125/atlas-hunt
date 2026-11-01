package ws

import (
	"errors"
	"testing"

	"atlas.sageyutta.com/internal/data"
	"atlas.sageyutta.com/internal/wsmessages"
)

func TestRegisterPlayer(t *testing.T) {
	room, _ := initializeFakeRoom(t)
	game := &fakeGame{}
	room.GameModel = game

	other := initializeFakeClient(t, room, "other", false)
	room.Players[other] = true

	joiner := initializeFakeClient(t, room, "joiner", false)

	room.registerPlayer(joiner)

	if !room.Players[joiner] {
		t.Error("joiner was not added to room.Players")
	}

	if got := game.getIncrementCalls(); len(got) != 1 || got[0] != FAKE_GAME_CODE {
		t.Errorf("IncrementPlayerCount calls = %v, want [%s]", got, FAKE_GAME_CODE)
	}

	// Joiner gets the welcome message, but not their own PlayerJoined broadcast.
	if msg := recvMsg(t, joiner); msg.Type != wsmessages.WelcomeToTheGame {
		t.Errorf("joiner got %v, want WelcomeToTheGame", msg.Type)
	}
	expectNoMsg(t, joiner)

	// The existing player is told about the new one.
	msg := recvMsg(t, other)
	if msg.Type != wsmessages.PlayerJoined {
		t.Fatalf("other got %v, want PlayerJoined", msg.Type)
	}
	if msg.Data["nickname"] != joiner.nickname || msg.Data["id"] != joiner.id {
		t.Errorf("PlayerJoined data = %v", msg.Data)
	}
}

func TestRegisterPlayer_IncrementErrorStillRegisters(t *testing.T) {
	room, _ := initializeFakeRoom(t)
	room.GameModel = &fakeGame{incrementErr: errors.New("redis down")}

	joiner := initializeFakeClient(t, room, "joiner", false)

	room.registerPlayer(joiner)

	if !room.Players[joiner] {
		t.Error("joiner should still be registered when redis fails")
	}
	if msg := recvMsg(t, joiner); msg.Type != wsmessages.WelcomeToTheGame {
		t.Errorf("got %v, want WelcomeToTheGame", msg.Type)
	}
}

func TestUnregisterPlayer(t *testing.T) {
	room, _ := initializeFakeRoom(t)
	game := &fakeGame{}
	player := &fakePlayer{}
	room.GameModel = game
	room.PlayerModel = player

	existingPlayer := initializeFakeClient(t, room, "existingPlayer", false)
	room.Players[existingPlayer] = true

	leavingPlayer := initializeFakeClient(t, room, "leavingPlayer", false)
	room.Players[leavingPlayer] = true

	if shouldClose := room.unregisterPlayer(leavingPlayer); shouldClose {
		t.Error("unregisterPlayer returned true, want false (existingPlayer is still online)")
	}

	if _, ok := room.Players[leavingPlayer]; ok {
		t.Error("leavingPlayer should have been deleted from room.Players")
	}

	if got := game.getDecrementCalls(); len(got) != 1 || got[0] != FAKE_GAME_CODE {
		t.Errorf("DecrementPlayerCount calls = %v, want [%s]", got, FAKE_GAME_CODE)
	}

	if got := game.getRemovePlayerCalls(); len(got) != 1 || got[0].code != FAKE_GAME_CODE || got[0].session != leavingPlayer.session {
		t.Errorf("RemovePlayerFromSet calls = %v, want code = %s; session = %s", got, FAKE_GAME_CODE, leavingPlayer.session)
	}

	if got := player.getDeleteCalls(); len(got) != 1 || got[0] != leavingPlayer.session {
		t.Errorf("player Delete calls = %v, want [%s]", got, leavingPlayer.session)
	}

	// existing player should get a PlayerLeft message
	msg := recvMsg(t, existingPlayer)
	if msg.Type != wsmessages.PlayerLeft {
		t.Fatalf("existingPlayer got %v, want PlayerLeft", msg.Type)
	}
	if msg.Data["id"] != leavingPlayer.id {
		t.Errorf("PlayerLeft id = %v, want %s", msg.Data["id"], leavingPlayer.id)
	}

	// the leaving player should get no messages
	expectClosed(t, leavingPlayer)
}

func TestUnregisterPlayer_LastPlayerLeaves(t *testing.T) {
	room, _ := initializeFakeRoom(t)
	game := &fakeGame{}
	playerModel := &fakePlayer{}
	room.GameModel = game
	room.PlayerModel = playerModel

	only := initializeFakeClient(t, room, "only", false)
	room.Players[only] = true

	if shouldClose := room.unregisterPlayer(only); !shouldClose {
		t.Error("unregisterPlayer returned false, want true when the room is empty")
	}

	if got := game.getRemovePlayerCalls(); len(got) != 0 {
		t.Errorf("RemovePlayerFromSet should not be called for the last player, got %v", got)
	}
	if got := playerModel.getDeleteCalls(); len(got) != 1 || got[0] != only.session {
		t.Errorf("player Delete calls = %v, want [%s]", got, only.session)
	}
	expectClosed(t, only)
}

func TestUnregisterPlayer_MidGameMarksOffline(t *testing.T) {
	room, _ := initializeFakeRoom(t)
	game := &fakeGame{}
	playerModel := &fakePlayer{}
	room.GameModel = game
	room.PlayerModel = playerModel
	room.State = data.GameStateInProgress

	stays := initializeFakeClient(t, room, "stays", false)
	room.Players[stays] = true
	leaves := initializeFakeClient(t, room, "leaves", false)
	room.Players[leaves] = true

	if shouldClose := room.unregisterPlayer(leaves); shouldClose {
		t.Error("unregisterPlayer returned true, want false")
	}

	online, ok := room.Players[leaves]
	if !ok || online {
		t.Errorf("leaves should remain in room.Players marked offline, got present=%v online=%v", ok, online)
	}
	if got := game.getRemovePlayerCalls(); len(got) != 0 {
		t.Errorf("RemovePlayerFromSet should not be called mid-game, got %v", got)
	}
	if got := playerModel.getDeleteCalls(); len(got) != 0 {
		t.Errorf("player Delete should not be called mid-game, got %v", got)
	}

	if msg := recvMsg(t, stays); msg.Type != wsmessages.PlayerLeft {
		t.Errorf("stays got %v, want PlayerLeft", msg.Type)
	}
	expectClosed(t, leaves)
}
