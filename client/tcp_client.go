package client

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	pb "card-game-client/proto/common"
	"card-game-client/util"

	"google.golang.org/protobuf/proto"
)

var (
	ErrTimeout = errors.New("request timeout")
	ErrClosed  = errors.New("connection closed")
)

type TCPClient struct {
	conn        net.Conn
	addr        string
	mu          sync.Mutex
	pendingReqs map[string]chan *pb.GameMessage
	done        chan struct{}
	onMessage   func(*pb.GameMessage)
}

func NewTCPClient(addr string) *TCPClient {
	return &TCPClient{
		addr:        addr,
		pendingReqs: make(map[string]chan *pb.GameMessage),
		done:        make(chan struct{}),
	}
}

func (c *TCPClient) Connect() error {
	conn, err := net.DialTimeout("tcp", c.addr, 5*time.Second)
	if err != nil {
		return err
	}
	c.conn = conn
	go c.readLoop()
	return nil
}

func (c *TCPClient) Close() {
	select {
	case <-c.done:
		// already closed
		return
	default:
		close(c.done)
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// 编码：varint32长度前缀 + protobuf消息
func encodeMessage(msg *pb.GameMessage) ([]byte, error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, binary.MaxVarintLen32+len(data))
	n := binary.PutUvarint(buf, uint64(len(data)))
	copy(buf[n:], data)
	return buf[:n+len(data)], nil
}

// 发送消息
func (c *TCPClient) Send(msg *pb.GameMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return ErrClosed
	}

	data, err := encodeMessage(msg)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(data)
	return err
}

// 发送并等待响应
func (c *TCPClient) SendAndWait(msg *pb.GameMessage, timeout time.Duration) (*pb.GameMessage, error) {
	seqID := util.GenerateUUID()
	if msg.Header == nil {
		msg.Header = &pb.MsgHeader{}
	}
	msg.Header.SeqId = seqID

	ch := make(chan *pb.GameMessage, 1)
	c.mu.Lock()
	c.pendingReqs[seqID] = ch
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.pendingReqs, seqID)
		c.mu.Unlock()
	}()

	if err := c.Send(msg); err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(timeout):
		return nil, ErrTimeout
	case <-c.done:
		return nil, ErrClosed
	}
}

// 读循环：解码消息并分发
func (c *TCPClient) readLoop() {
	for {
		select {
		case <-c.done:
			return
		default:
		}

		// 读取消息长度
		length, err := binary.ReadUvarint(&byteReader{c.conn})
		if err != nil {
			return
		}

		// 读取消息体
		buf := make([]byte, length)
		if _, err := io.ReadFull(c.conn, buf); err != nil {
			return
		}

		// 解码 protobuf
		msg := &pb.GameMessage{}
		if err := proto.Unmarshal(buf, msg); err != nil {
			continue
		}

		// 处理响应（如果有 SeqId）
		if msg.Header != nil && msg.Header.SeqId != "" {
			c.mu.Lock()
			if ch, ok := c.pendingReqs[msg.Header.SeqId]; ok {
				ch <- msg
			}
			c.mu.Unlock()
		}

		// 调用消息回调
		if c.onMessage != nil {
			c.onMessage(msg)
		}
	}
}

// 判断是否已连接
func (c *TCPClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil
}

// 设置消息回调
func (c *TCPClient) SetOnMessageHandler(handler func(*pb.GameMessage)) {
	c.onMessage = handler
}

// 获取连接（用于调试）
func (c *TCPClient) GetConn() net.Conn {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn
}

type byteReader struct {
	r io.Reader
}

func (br *byteReader) ReadByte() (byte, error) {
	var buf [1]byte
	_, err := br.r.Read(buf[:])
	return buf[0], err
}
