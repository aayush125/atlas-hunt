package main

import (
	"net/http"

	"atlas.sageyutta.com/internal/cc_context"
)

func (app *application) leaveGameHandler(w http.ResponseWriter, r *http.Request) {
	token := cc_context.ContextGetSessionToken(r)
	player := cc_context.ContextGetPlayer(r)

	err := app.models.Game.RemovePlayerFromSet(player.GameID, player.SessionToken)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Player.Delete(token)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}
