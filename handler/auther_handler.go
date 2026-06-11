package handler

import (
	"card-game-client/client"
	aurhor "card-game-client/proto/author"
	pb "card-game-client/proto/common"
	"card-game-client/util"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
)

type AuthorManager struct {
	client   *client.TCPClient
	stopCh   chan struct{}
	playerID string
}

func NewAuthorManager(cli *client.TCPClient) *AuthorManager {
	return &AuthorManager{
		client: cli,
		stopCh: make(chan struct{}),
	}
}

// 注册
func (a *AuthorManager) Register() error {

	register := &aurhor.RegisterRequest{
		Username:   "test2",
		Password:   "123456",
		Nickname:   "test",
		ClientType: 1,
	}
	// 将登录请求数据序列化为字节流
	bodyBytes, err := proto.Marshal(register)
	if err != nil {
		return fmt.Errorf("marshal login request failed: %w", err)
	}
	req := &pb.GameMessage{
		Header: &pb.MsgHeader{
			MsgId:     util.GenerateUUID(),
			Timestamp: time.Now().UnixMilli(),
		},
		MessageType: pb.MessageType_AUTH_REGISTER_REQUEST,
		Body:        bodyBytes,
	}

	// 发送登录请求并等待响应，设置5秒超时时间
	resp, err := a.client.SendAndWait(req, 5*time.Second)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	// 反序列化响应数据为登录响应对象
	loginRsp := &aurhor.RegisterResponse{}
	if err := proto.Unmarshal(resp.GetBody(), loginRsp); err != nil {
		return fmt.Errorf("unmarshal response failed: %w", err)
	}

	// 检查登录响应的状态码，如果不是成功则返回错误
	if loginRsp.GetCode() != pb.Code_SUCCESS {
		return fmt.Errorf("login failed: code=%d, msg=%s",
			loginRsp.GetCode(), loginRsp.GetMessage())
	}

	a.playerID = loginRsp.GetPlayerId()
	// 打印登录成功的日志，只显示token的前20个字符
	fmt.Printf("Login success: playerId=%s, token=%s\n",
		a.playerID, "...")
	return nil
}
