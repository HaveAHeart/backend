package model

import (
	"sync"
	"time"

	pkggame "github.com/renju24/backend/pkg/game"
)

type Move struct {
	GameID      int64
	UserID      int64
	XCoordinate int
	YCoordinate int
}

type GameStatus int

const (
	WaitingOpponent GameStatus = iota
	InProgress
	Finished
)

type Game struct {
	ID          int64      `json:"id"`
	BlackUserID int64      `json:"black_user_id"`
	WhiteUserID int64      `json:"white_user_id"`
	Winner      *int64     `json:"winner_id"`
	StartedAt   time.Time  `json:"started_at"`
	Status      GameStatus `json:"status"`
	FinishedAt  *time.Time `json:"finished_at"`

	mu   sync.Mutex
	game *pkggame.Game
}

// ApplyMove ...
func (g *Game) ApplyMove(userID int64, x, y int) (winner pkggame.Color, err error) {
	g.mu.Lock()
	winner, err = g.game.ApplyMove(pkggame.NewMove(x, y, g.GetColorByUserID(userID)))
	g.mu.Unlock()
	return
}

// GetColorByUserID ...
func (g *Game) GetColorByUserID(userID int64) pkggame.Color {
	switch userID {
	case g.BlackUserID:
		return pkggame.Black
	case g.WhiteUserID:
		return pkggame.White
	}
	return pkggame.Nil
}

// GetUserIDByColor ...
func (g *Game) GetUserIDByColor(color pkggame.Color) int64 {
	switch color {
	case pkggame.Black:
		return g.BlackUserID
	case pkggame.White:
		return g.WhiteUserID
	}
	return 0
}

type GameHistoryItem struct {
	ID             int64   `json:"id"`
	BlackUsername  string  `json:"black_username"`
	WhiteUsername  string  `json:"white_username"`
	WinnerUsername *string `json:"winner_username"`
}
