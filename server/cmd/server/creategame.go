package main

import (
	"net/http"
	"time"

	"game.crisis-control.com/internal/data"
	"game.crisis-control.com/internal/gameauth"
	"game.crisis-control.com/internal/validator"
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
		NumberOfPlayers: input.NumberOfPlayers,
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

	token, err := gameauth.GenerateSessionToken()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	gameRequest.GameID = gameID.Plaintext
	gameRequest.CreatedAt = time.Now()

	err = app.models.Game.Insert(&gameRequest)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	user_uuid, err := uuid.NewRandom()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	player := data.Player{
		Nickname:            gameRequest.CreatorNickname,
		GameID:              gameRequest.GameID,
		JoinTimestamp:       time.Now(),
		LastRejoinTimestamp: time.Now(),
		PlayerID:            user_uuid,
		SessionToken:        token,
	}

	err = app.models.Player.Insert(&player)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Game.AddPlayerToGame(gameRequest.GameID)
	if err != nil {
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

	err = app.writeJSON(w, http.StatusCreated, envelope{"game": gameRequest}, nil, cookie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
