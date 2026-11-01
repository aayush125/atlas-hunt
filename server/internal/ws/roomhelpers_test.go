package ws

import "testing"

func TestGetPlayerWithSession(t *testing.T) {
	room, exitCh := initializeFakeRoom(t)
	gameModel := &fakeGame{}
	playerModel := &fakePlayer{}
	room.GameModel = gameModel
	room.PlayerModel = playerModel

	player := initializeFakeClient(t, room, "player", false)
	room.Players[player] = true

	serverConn, _ := newWebsocketPair(t)
	player.conn = serverConn

	go room.Run()

	client := room.GetPlayerClientWithSessionToken(player.session)
	if client != player {
		t.Errorf("Got client %v; want %v", client, player)
	}

	room.Shutdown <- struct{}{}
	room.Unregister <- player
	<-exitCh

	if got := gameModel.getAllPlayerCalls(); len(got) != 1 || got[0] != FAKE_GAME_CODE {
		t.Errorf("GetAllPlayers calls = %v, want [%s]", got, FAKE_GAME_CODE)
	}

	if got := playerModel.getDeleteCalls(); len(got) != 0 {
		t.Errorf("player Delete calls = %v, want []", got)
	}
}
