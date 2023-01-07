package apiserver

import (
	oauth "github.com/renju24/backend/internal/pkg/oauth2"
	"github.com/renju24/backend/model"
)

type Database interface {
	// Create new user.
	CreateUser(username, email, passwordBcrypt string) (*model.User, error)

	// Create new user from oauth.
	CreateUserOauth(username string, email *string, oauthID string, oauthSerivce oauth.Service) (*model.User, error)

	// Get user by login.
	GetUserByLogin(login string) (*model.User, error)

	// Get user by ID.
	GetUserByID(userID int64) (*model.User, error)

	// Create new game.
	CreateGame(blackUserID, whiteUserID int64) (gameID int64, err error)

	// Get game by id.
	GetGameByID(gameID int64) (*model.Game, error)

	// Get game moves by id.
	GetGameMovesByID(gameID int64) ([]model.Move, error)

	// Create new move.
	CreateMove(gameID, userID int64, x, y int) error

	// Is user a game member?
	IsGameMember(userID, gameID int64) (bool, error)

	// Is user playing a game right now?
	IsPlaying(userID int64) (bool, error)

	// Find users by username.
	FindUsers(username string) ([]*model.User, error)

	// Get game history by username.
	GameHistory(username string) ([]model.GameHistoryItem, error)

	// Top10 return the top 10 users by ranking.
	Top10() ([]*model.User, error)

	// Delete a game from database.
	DeclineGameInvitation(userID int64, gameID int64) error

	// Set game status to InProgress.
	StartGame(gameID int64) error

	// Set game status to Finished.
	FinishGameWithWinner(gameID, winnerID int64) error

	// Close
	Close() error
}
