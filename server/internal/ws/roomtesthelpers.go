package ws

import (
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"
)

const (
	FAKE_GAME_CODE   = "FAKEGM"
	FAKE_MAX_PLAYERS = 4

	FAKE_ROUND_TIME  = 500 * time.Millisecond
	FAKE_BACKUP_TIME = 1 * time.Second
	FAKE_TTL_RESET   = 1 * time.Second
)

type removePlayerCallData struct {
	code    string
	session string
}

type fakeGame struct {
	Game
	mu sync.Mutex

	incrementCalls []string
	incrementErr   error

	decrementCalls []string
	decrementErr   error

	removePlayerCalls []removePlayerCallData
	removePlayerErr   error

	allPlayerCalls    []string
	allPlayerCallsErr error

	cleanupGameCalls    []string
	cleanupGameCallsErr error
}

func (f *fakeGame) IncrementPlayerCount(code string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.incrementCalls = append(f.incrementCalls, code)
	return f.incrementErr
}

func (f *fakeGame) DecrementPlayerCount(code string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.decrementCalls = append(f.decrementCalls, code)
	return f.decrementErr
}

func (f *fakeGame) RemovePlayerFromSet(code, session string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.removePlayerCalls = append(f.removePlayerCalls, removePlayerCallData{code: code, session: session})
	return f.removePlayerErr
}

func (f *fakeGame) GetAllPlayers(code string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.allPlayerCalls = append(f.allPlayerCalls, code)
	return []string{}, f.allPlayerCallsErr
}

func (f *fakeGame) CleanupGame(code string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cleanupGameCalls = append(f.cleanupGameCalls, code)
	return f.cleanupGameCallsErr
}

func (f *fakeGame) getIncrementCalls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.incrementCalls...)
}

func (f *fakeGame) getDecrementCalls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.decrementCalls...)
}

func (f *fakeGame) getRemovePlayerCalls() []removePlayerCallData {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]removePlayerCallData(nil), f.removePlayerCalls...)
}

func (f *fakeGame) getAllPlayerCalls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.allPlayerCalls...)
}

type fakePlayer struct {
	Player
	mu sync.Mutex

	deleteCalls []string
	deleteErr   error
}

func (f *fakePlayer) Delete(session string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.deleteCalls = append(f.deleteCalls, session)
	return f.deleteErr
}

func (f *fakePlayer) getDeleteCalls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.deleteCalls...)
}

func initializeFakeRoom(t testing.TB) (*Room, <-chan string) {
	t.Helper()

	exitCh := make(chan string, 1)
	onExit := func(code string) { exitCh <- code }

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	room := NewRoom(
		FAKE_GAME_CODE,
		FAKE_MAX_PLAYERS,
		&fakeGame{},
		&fakePlayer{},
		logger,
		nil,
		onExit,
		FAKE_ROUND_TIME,
		FAKE_BACKUP_TIME,
		FAKE_TTL_RESET,
	)

	t.Cleanup(func() {
		room.timer.Stop()
		room.backupTimer.Stop()
	})

	return room, exitCh
}
