package apiserver

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/armantarkhanian/websocket"
	"github.com/centrifugal/centrifuge"
	"github.com/renju24/backend/internal/pkg/apierror"
)

func (*APIServer) OnAlive(*websocket.Client) {}

func (*APIServer) OnDisconect(*websocket.Client, centrifuge.DisconnectEvent) {}

func (app *APIServer) OnSubscribe(c *websocket.Client, e centrifuge.SubscribeEvent) (centrifuge.SubscribeReply, error) {
	if strings.HasPrefix(e.Channel, "user_") {
		if e.Channel != "user_"+c.UserID() {
			return centrifuge.SubscribeReply{}, centrifuge.ErrorPermissionDenied
		}
	}

	if strings.HasPrefix(e.Channel, "game_") {
		gameID, err := strconv.ParseInt(e.Channel[5:], 10, 64)
		if err != nil {
			return centrifuge.SubscribeReply{}, centrifuge.ErrorBadRequest
		}
		userID, err := strconv.ParseInt(c.UserID(), 10, 64)
		if err != nil {
			return centrifuge.SubscribeReply{}, centrifuge.ErrorBadRequest
		}

		// Check if the user is a game member.
		ok, err := app.db.IsGameMember(gameID, userID)
		if err != nil {
			return centrifuge.SubscribeReply{}, centrifuge.ErrorInternal
		}
		if !ok {
			return centrifuge.SubscribeReply{}, centrifuge.ErrorPermissionDenied
		}
	}

	return centrifuge.SubscribeReply{}, nil
}

func (apiServer *APIServer) OnRPC(c *websocket.Client, rpc centrifuge.RPCEvent) (centrifuge.RPCReply, error) {
	if !c.Authorized() {
		return centrifuge.RPCReply{}, apierror.ErrorUnauthorized
	}
	var response any
	var err error
	switch rpc.Method {
	case "get_user":
		response, err = apiServer.GetUser(c, rpc.Data)
	case "game_history":
		response, err = apiServer.GameHistory(c, rpc.Data)
	case "find_users":
		response, err = apiServer.FindUsers(c, rpc.Data)
	case "call_for_game":
		response, err = apiServer.CallForGame(c, rpc.Data)
	case "top_10":
		response, err = apiServer.Top10(c, rpc.Data)
	case "decline_game_invitation":
		response, err = apiServer.DeclineGameInvitation(c, rpc.Data)
	case "accept_game_invitation":
		response, err = apiServer.AcceptGameInvitation(c, rpc.Data)
	case "make_move":
		response, err = apiServer.MakeMove(c, rpc.Data)
	default:
		return centrifuge.RPCReply{}, centrifuge.ErrorMethodNotFound
	}
	if err != nil {
		return centrifuge.RPCReply{}, err
	}

	data, err := json.Marshal(response)
	if err != nil {
		apiServer.logger.Error().Err(err).Send()
		return centrifuge.RPCReply{}, centrifuge.ErrorInternal
	}

	return centrifuge.RPCReply{Data: data}, nil
}

func (*APIServer) OnUnsubscribe(*websocket.Client, centrifuge.UnsubscribeEvent) {}

func (*APIServer) OnPublish(*websocket.Client, centrifuge.PublishEvent) (centrifuge.PublishReply, error) {
	return centrifuge.PublishReply{}, nil
}

func (*APIServer) OnRefresh(*websocket.Client, centrifuge.RefreshEvent) (centrifuge.RefreshReply, error) {
	return centrifuge.RefreshReply{}, nil
}

func (*APIServer) OnSubRefresh(*websocket.Client, centrifuge.SubRefreshEvent) (centrifuge.SubRefreshReply, error) {
	return centrifuge.SubRefreshReply{}, nil
}

func (*APIServer) OnMessage(*websocket.Client, centrifuge.MessageEvent) {}

func (*APIServer) OnPresence(*websocket.Client, centrifuge.PresenceEvent) (centrifuge.PresenceReply, error) {
	return centrifuge.PresenceReply{}, nil
}

func (*APIServer) OnPresenceStats(*websocket.Client, centrifuge.PresenceStatsEvent) (centrifuge.PresenceStatsReply, error) {
	return centrifuge.PresenceStatsReply{}, nil
}

func (*APIServer) OnHistory(*websocket.Client, centrifuge.HistoryEvent) (centrifuge.HistoryReply, error) {
	return centrifuge.HistoryReply{}, nil
}
