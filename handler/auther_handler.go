package handler

import (
	"card-game-client/client"
	"fmt"
	"time"

	pb "card-game-client/proto/common"
	gatewaypb "card-game-client/proto/gateway"

	"google.golang.org/protobuf/proto"
)

type LoginHandler struct {
	client   *client.TCPClient
	token    string
	playerID string
}

func NewLoginHandler(cli *client.TCPClient) *LoginHandler {
	return &LoginHandler{client: cli}
}

func (h *LoginHandler) Login(username, password string) error {
	loginReq := &gatewaypb.LoginReq{
		Username:      username,
		Password:      password,
		ClientType:    1,
		ClientVersion: "1.0.0",
	}

	bodyBytes, err := proto.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("marshal login request failed: %w", err)
	}

	req := &pb.GameMessage{
		Header: &pb.MsgHeader{
			MsgId:     "1001",
			Timestamp: time.Now().UnixMilli(),
		},
		Body:        bodyBytes,
		MessageType: pb.MessageType_AUTH_LOGIN_REQUEST,
	}

	resp, err := h.client.SendAndWait(req, 5*time.Second)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	loginRsp := &gatewaypb.LoginRsp{}
	if err := proto.Unmarshal(resp.GetBody(), loginRsp); err != nil {
		return fmt.Errorf("unmarshal response failed: %w", err)
	}

	if loginRsp.GetCode() != pb.Code_SUCCESS {
		return fmt.Errorf("login failed: code=%d, msg=%s",
			loginRsp.GetCode(), loginRsp.GetMessage())
	}

	h.token = loginRsp.GetToken()
	h.playerID = loginRsp.GetPlayerInfo().GetPlayerId()
	fmt.Printf("Login success: playerId=%s, token=%s\n",
		h.playerID, h.token[:20]+"...")
	return nil
}

func (h *LoginHandler) GetToken() string    { return h.token }
func (h *LoginHandler) GetPlayerID() string { return h.playerID }
