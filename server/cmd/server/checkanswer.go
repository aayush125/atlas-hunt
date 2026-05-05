package main

import (
	"errors"
	"net/http"

	"game.crisis-control.com/internal/gamedata"
)

func (app *application) checkTriviaAnswer(w http.ResponseWriter, r *http.Request) {
	var input struct {
		QuestionID     string `json:"questionId"`
		SelectedAnswer string `json:"selectedAnswer"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	correct, err := gamedata.CheckAnswer(input.QuestionID, input.SelectedAnswer)
	if err != nil {
		switch {
		case errors.Is(err, gamedata.ErrQuestionNotFound):
			app.badRequestResponse(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"correct": correct}, nil, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
