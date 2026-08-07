package main

import (
	"errors"
	"fmt"
	"net/http"

	"atlas.sageyutta.com/internal/cc_context"
	"atlas.sageyutta.com/internal/data"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")

				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireSessionToken(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			switch err {
			case http.ErrNoCookie:
				app.noSessionTokenResponse(w, r)
			default:
				app.badRequestResponse(w, r, err)
			}
			return
		}

		token := cookie.Value
		if token == "" {
			app.invalidSessionTokenResponse(w, r)
			return
		}

		player, err := app.models.Player.GetForSessionToken(token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidSessionTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = cc_context.ContextSetSessionToken(r, token)
		r = cc_context.ContextSetPlayer(r, player)
		next.ServeHTTP(w, r)
	})
}
