package main

import (
	"fmt"
	"net/http"
	"time"

	"game.crisis-control.com/internal/data"
	"game.crisis-control.com/internal/gameauth"
	"game.crisis-control.com/internal/validator"
	"game.crisis-control.com/internal/ws"
	"github.com/google/uuid"
)

func (app *application) createGameHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Nickname        string `json:"nickname"`
		NumberOfPlayers int    `json:"number_of_players"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	gameRequest := data.Game{
		CreatorNickname: input.Nickname,
		MaxPlayerCount:  input.NumberOfPlayers,
	}

	v := validator.New()

	if data.ValidateCreateGame(v, &gameRequest); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	gameID, err := gameauth.GenerateGameID()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.mu.Lock()
	_, ok := app.gameRooms[gameID.Plaintext]
	app.mu.Unlock()

	if ok {
		app.serverErrorResponse(w, r, fmt.Errorf("duplicate room"))
		return
	}

	token, err := gameauth.GenerateSessionToken()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	gameRequest.GameID = gameID.Plaintext
	gameRequest.CreatedAt = time.Now()
	gameRequest.State = data.GameStateLobby
	gameRequest.CurrentMap = data.GameMapNone

	err = app.models.Game.Insert(&gameRequest)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	userUUID, err := uuid.NewRandom()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	player := data.Player{
		Nickname:            gameRequest.CreatorNickname,
		GameID:              gameRequest.GameID,
		JoinTimestamp:       time.Now(),
		LastRejoinTimestamp: time.Now(),
		PlayerID:            userUUID,
		IsCreator:           true,
		SessionToken:        token,
	}

	err = app.models.Player.Insert(&player)
	if err != nil {
		app.models.Game.Delete(gameRequest.GameID)
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Game.AddPlayerToSet(gameRequest.GameID, player.SessionToken)
	if err != nil {
		app.models.Game.Delete(gameRequest.GameID)
		app.models.Player.Delete(player.SessionToken)
		app.serverErrorResponse(w, r, err)
		return
	}

	// Set initial TTL of 2 hours
	err = app.models.Game.SetTTLForPlayerSet(gameRequest.GameID, 2*time.Hour)
	if err != nil {
		app.models.Game.Delete(gameRequest.GameID)
		app.models.Player.Delete(player.SessionToken)
		app.models.Game.RemovePlayerFromSet(gameRequest.GameID, player.SessionToken)
		app.serverErrorResponse(w, r, err)
		return
	}

	room := ws.NewRoom(app.models, gameRequest.GameID, gameRequest.MaxPlayerCount, app.models.Game, app.models.Player, app.logger, app.engine)

	app.mu.Lock()
	if _, ok := app.gameRooms[gameRequest.GameID]; ok {
		app.mu.Unlock()
		app.serverErrorResponse(w, r, fmt.Errorf("duplicate room"))
		return
	}
	app.gameRooms[gameRequest.GameID] = room
	app.mu.Unlock()

	app.background(room.Run)

	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(30 * time.Minute),
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"game": gameRequest, "player": player}, nil, cookie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
