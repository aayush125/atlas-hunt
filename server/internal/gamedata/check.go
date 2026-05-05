package gamedata

import "errors"

var ErrQuestionNotFound = errors.New("question not found")

func CheckAnswer(questionID, answer string) (bool, error) {
	correct, ok := answers[questionID]
	if !ok {
		return false, ErrQuestionNotFound
	}
	return answer == correct, nil
}
