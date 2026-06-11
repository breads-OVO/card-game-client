package handler

import (
	"time"

	"card-game-client/client"
	pb "card-game-client/proto/common"
	"card-game-client/util"
)

type HeartbeatManager struct {
	client   *client.TCPClient
	interval time.Duration
	stopCh   chan struct{}
}

func NewHeartbeatManager(cli *client.TCPClient, interval time.Duration) *HeartbeatManager {
	return &HeartbeatManager{
		client:   cli,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (h *HeartbeatManager) Start() {
	go func() {
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.sendHeartbeat()
			case <-h.stopCh:
				return
			}
		}
	}()
}

func (h *HeartbeatManager) Stop() {
	close(h.stopCh)
}

func (h *HeartbeatManager) sendHeartbeat() {
	req := &pb.GameMessage{
		Header: &pb.MsgHeader{
			MsgId:     util.GenerateUUID(),
			Timestamp: time.Now().UnixMilli(),
		},
		MessageType: pb.MessageType_HEARTBEAT,
	}

	// 异步发送，不等待响应
	_ = h.client.Send(req)
}
