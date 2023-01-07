package apiserver

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/armantarkhanian/websocket"
	"github.com/renju24/backend/internal/pkg/apierror"
)

type RPCDeclineGameInvitationRequest struct {
	GameID int64 `json:"game_id"`
}

type RPCDeclineGameInvitationResponse struct{}

func (apiServer *APIServer) DeclineGameInvitation(c *websocket.Client, jsonData []byte) (*RPCDeclineGameInvitationResponse, error) {
	var req RPCDeclineGameInvitationRequest
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
	if err = apiServer.db.DeclineGameInvitation(opponentID, req.GameID); err != nil {
		apiServer.logger.Error().Err(err).Send()
		return nil, apierror.ErrorInternal
	}
	if _, err = apiServer.PublishEvent(fmt.Sprintf("game_%d", req.GameID), &EventDeclineGameInvitation{}); err != nil {
		apiServer.logger.Error().Err(err).Send()
		return nil, apierror.ErrorInternal
	}
	return &RPCDeclineGameInvitationResponse{}, nil
}
