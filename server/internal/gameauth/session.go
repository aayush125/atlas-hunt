package gameauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
)

type GameID struct {
	Plaintext string
	Hash      []byte
}

func GenerateGameID() (*GameID, error) {
	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	plainTextGameID := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)[:6]
	hash := sha256.Sum256([]byte(plainTextGameID))

	gameID := &GameID{
		Plaintext: plainTextGameID,
		Hash:      hash[:],
	}

	return gameID, nil
}

func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
