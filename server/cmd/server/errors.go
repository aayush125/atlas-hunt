package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := app.constructReturnReponse("server_error", "The server encountered a problem and could not process your request")
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := app.constructReturnReponse("not_found", "The requested resource could not be found")
	app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *application) methoNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := app.constructReturnReponse("method_not_allowed", fmt.Sprintf("the %s method is not supported for this resource", r.Method))
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	message := app.constructReturnReponse("validation_failed", errors)
	app.errorResponse(w, r, http.StatusUnprocessableEntity, message)
}

func (app *application) noSessionTokenResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusUnauthorized, envelope{"unauthorized": "no session token"})
}

func (app *application) invalidSessionTokenResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusUnauthorized, envelope{"unauthorized": "invalid session token"})
}

func (app *application) roomNotFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotFound, envelope{"message": "the requested game room does not exist"})
}

func (app *application) roomFullResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusConflict, envelope{"message": "the requested game room is already full"})
}

func (app *application) treasureNotFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotAcceptable, envelope{"message": "the requested treasure could not be found"})
}

func (app *application) customErrorResponse(w http.ResponseWriter, r *http.Request, status int, err error) {
	app.errorResponse(w, r, status, envelope{"message": err.Error()})
}
