package cc_context

import (
	"context"
	"net/http"

	"game.crisis-control.com/internal/data"
)

type contextKey string

const sessionContextKey = contextKey("session_token")
const playerContextKey = contextKey("player")

func ContextSetSessionToken(r *http.Request, token string) *http.Request {
	ctx := context.WithValue(r.Context(), sessionContextKey, token)
	return r.WithContext(ctx)
}

func ContextGetSessionToken(r *http.Request) string {
	token, ok := r.Context().Value(sessionContextKey).(string)
	if !ok {
		panic("missing session token in request context")
	}

	return token
}

func ContextSetPlayer(r *http.Request, player *data.Player) *http.Request {
	ctx := context.WithValue(r.Context(), playerContextKey, player)
	return r.WithContext(ctx)
}

func ContextGetPlayer(r *http.Request) *data.Player {
	player, ok := r.Context().Value(playerContextKey).(*data.Player)
	if !ok {
		panic("missing player in request context")
	}

	return player
}
