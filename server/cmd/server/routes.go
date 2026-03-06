package main

import (
	"net/http"

	"net/http/pprof"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	// custom errors
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methoNotAllowedResponse)

	// routes
	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/game/create", app.createGameHandler)
	router.HandlerFunc(http.MethodPost, "/game/join", app.registerPlayerHandler)
	router.HandlerFunc(http.MethodGet, "/ws", app.requireSessionToken(app.websocketHandler))

	// profiler
	if app.config.env == "development" {
		router.HandlerFunc(http.MethodGet, "/debug/pprof/", pprof.Index)
		router.HandlerFunc(http.MethodGet, "/debug/pprof/cmdline", pprof.Cmdline)
		router.HandlerFunc(http.MethodGet, "/debug/pprof/profile", pprof.Profile)
		router.HandlerFunc(http.MethodGet, "/debug/pprof/symbol", pprof.Symbol)
		router.HandlerFunc(http.MethodGet, "/debug/pprof/trace", pprof.Trace)
		router.Handler(http.MethodGet, "/debug/pprof/allocs", pprof.Handler("allocs"))
		router.Handler(http.MethodGet, "/debug/pprof/goroutine", pprof.Handler("goroutine"))
		router.Handler(http.MethodGet, "/debug/pprof/heap", pprof.Handler("heap"))
		router.Handler(http.MethodGet, "/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
		router.Handler(http.MethodGet, "/debug/pprof/block", pprof.Handler("block"))
	}

	return app.recoverPanic(router)
}
