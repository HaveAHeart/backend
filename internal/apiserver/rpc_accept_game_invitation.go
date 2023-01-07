package apiserver

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/armantarkhanian/websocket"
	"github.com/renju24/backend/internal/pkg/apierror"
)

type RPCAcceptGameInvitationRequest struct {
	GameID int64 `json:"game_id"`
}

type RPCAcceptGameInvitationResponse struct{}

func (apiServer *APIServer) AcceptGameInvitation(c *websocket.Client, jsonData []byte) (*RPCAcceptGameInvitationResponse, error) {
	var req RPCAcceptGameInvitationRequest
	if err := json.Unmarshal(jsonData, &req); err != nil {
		return nil, apierror.ErrorBadRequest
	}
	opponentID, err := strconv.ParseInt(c.UserID(), 10, 64)
	if err != nil {
		return nil, apierror.ErrorUnauthorized
	}
	// Check if user is a game member.
	isGameMember, err := apiServer.db.IsGameMember(opponentID, req.GameID)
	if err != nil {
		apiServer.logger.Error().Err(err).Send()
		return nil, apierror.ErrorInternal
	}
	if !isGameMember {
		return nil, apierror.ErrorPermissionDenied
	}
	// Change game status and started_at in database.
	if err = apiServer.db.StartGame(req.GameID); err != nil {
		apiServer.logger.Error().Err(err).Send()
		return nil, apierror.ErrorInternal
	}
	// Publish event that game is started.
	if _, err = apiServer.PublishEvent(fmt.Sprintf("game_%d", req.GameID), &EventGameStarted{}); err != nil {
		apiServer.logger.Error().Err(err).Send()
		return nil, apierror.ErrorInternal
	}
	return &RPCAcceptGameInvitationResponse{}, nil
}
