package main

import (
	"errors"
	"net/http"
	"time"

	"game.crisis-control.com/internal/data"
	"game.crisis-control.com/internal/gameauth"
	"game.crisis-control.com/internal/validator"
	"github.com/google/uuid"
)

func (app *application) registerPlayerHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Nickname string `json:"nickname"`
		GameCode string `json:"game_code"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	joinRequest := data.JoinGame{
		Nickname: input.Nickname,
		GameCode: input.GameCode,
	}

	v := validator.New()

	if data.ValidateJoinGame(v, &joinRequest); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	game, err := app.models.Game.GetForCode(joinRequest.GameCode)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.roomNotFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if game.State != data.GameStateLobby || game.CurrentPlayerCount >= game.MaxPlayerCount {
		app.roomFullResponse(w, r)
		return
	}

	var existingPlayers []data.PlayerInfo
	allPlayers, err := app.models.Game.GetAllPlayers(joinRequest.GameCode)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	for _, token := range allPlayers {
		info, err := app.models.Player.GetForSessionToken(token)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		existingPlayers = append(existingPlayers, data.PlayerInfo{
			Nickname: info.Nickname,
			ID:       info.PlayerID.String(),
		})
	}

	token, err := gameauth.GenerateSessionToken()
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
		Nickname:            joinRequest.Nickname,
		GameID:              joinRequest.GameCode,
		JoinTimestamp:       time.Now(),
		LastRejoinTimestamp: time.Now(),
		PlayerID:            userUUID,
		IsCreator:           false,
		SessionToken:        token,
	}

	err = app.models.Player.Insert(&player)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Game.AddPlayerToSet(player.GameID, player.SessionToken)
	if err != nil {
		app.models.Player.Delete(player.SessionToken)
		app.serverErrorResponse(w, r, err)
		return
	}

	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(30 * time.Minute),
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"player": player, "game": game, "existing_players": existingPlayers}, nil, cookie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
