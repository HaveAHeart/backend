package apiserver

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/armantarkhanian/websocket"
	"github.com/renju24/backend/internal/pkg/apierror"
	"github.com/renju24/backend/model"
	pkggame "github.com/renju24/backend/pkg/game"
)

type RPCMakeMoveRequest struct {
	GameID      int64 `json:"game_id"`
	XCoordinate int   `json:"x_coordinate"`
	YCoordinate int   `json:"y_coordinate"`
}

type RPCMakeMoveResponse struct{}

func (apiServer *APIServer) MakeMove(c *websocket.Client, jsonData []byte) (*RPCMakeMoveResponse, error) {
	var req RPCMakeMoveRequest
	if err := json.Unmarshal(jsonData, &req); err != nil {
		return nil, apierror.ErrorBadRequest
	}
	userID, err := strconv.ParseInt(c.UserID(), 10, 64)
	if err != nil {
		return nil, apierror.ErrorUnauthorized
	}
	// Check game in database.
	game, err := apiServer.db.GetGameByID(req.GameID)
	if err != nil {
		apiServer.logger.Error().Err(err).Send()
		return nil, apierror.ErrorInternal
	}
	if game.Status != model.InProgress {
		apiServer.logger.Error().Err(err).Send()
		return nil, apierror.ErrorGameIsNotActive
	}
	// Check if user is a game member.
	isGameMember, err := apiServer.db.IsGameMember(userID, req.GameID)
	if err != nil {
		apiServer.logger.Error().Err(err).Send()
		return nil, apierror.ErrorInternal
	}
	if !isGameMember {
		return nil, apierror.ErrorPermissionDenied
	}
	// Retrieve previous game moves from database and apply them all.
	moves, err := apiServer.db.GetGameMovesByID(req.GameID)
	if err != nil {
		apiServer.logger.Error().Err(err).Send()
		return nil, apierror.ErrorInternal
	}
	for _, move := range moves {
		if _, err = game.ApplyMove(move.UserID, move.XCoordinate, move.YCoordinate); err != nil {
			return nil, err
		}
	}
	// Apply next move.
	winnerColor, err := game.ApplyMove(userID, req.XCoordinate, req.YCoordinate)
	if err != nil {
		return nil, err
	}
	// If success then add move into database.
	if err = apiServer.db.CreateMove(req.GameID, userID, req.XCoordinate, req.YCoordinate); err != nil {
		apiServer.logger.Error().Err(err).Send()
		return nil, apierror.ErrorInternal
	}
	// Publish move into game's channel.
	gameChannel := fmt.Sprintf("game_%d", game.ID)
	var event Event
	switch winnerColor {
	case pkggame.Nil:
		event = &EventMove{
			UserID:      userID,
			XCoordinate: req.XCoordinate,
			YCoordinate: req.YCoordinate,
		}
	default:
		// Finish game if there is a winner.
		winnerID := game.GetUserIDByColor(winnerColor)
		event = &EventGameEndedWithWinner{
			WinnerID: winnerID,
		}
		if err = apiServer.db.FinishGameWithWinner(req.GameID, winnerID); err != nil {
			apiServer.logger.Error().Err(err).Send()
			return nil, apierror.ErrorInternal
		}
	}
	if _, err = apiServer.PublishEvent(gameChannel, event); err != nil {
		apiServer.logger.Error().Err(err).Send()
		return nil, apierror.ErrorInternal
	}
	return &RPCMakeMoveResponse{}, nil
}

var (
	mu    sync.RWMutex          // Mutex that protects all games.
	games map[int64]*model.Game // Active games by their ID.
)
