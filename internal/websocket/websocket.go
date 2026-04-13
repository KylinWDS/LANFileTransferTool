package websocket

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"lanfiletransfertool/internal/stats"
	"lanfiletransfertool/pkg/logger"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  64 * 1024,
	WriteBufferSize: 64 * 1024,
}

type MessageType int

const (
	MsgHandshake MessageType = iota
	MsgFileRequest
	MsgFileData
	MsgFileEnd
	MsgError
	MsgProgress
)

type Message struct {
	Type     MessageType `json:"type"`
	FileID   string      `json:"file_id,omitempty"`
	FileName string      `json:"file_name,omitempty"`
	FileSize int64       `json:"file_size,omitempty"`
	Offset   int64       `json:"offset,omitempty"`
	ChunkLen int         `json:"chunk_len,omitempty"`
	Data     []byte      `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
	Progress float64     `json:"progress,omitempty"`
}

type Session struct {
	ID          string
	Conn        *websocket.Conn
	FileName    string
	FileSize    int64
	Transferred int64
	StartTime   time.Time
	Status      string
}

type Service struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewService() *Service {
	return &Service{
		sessions: make(map[string]*Session),
	}
}

func (s *Service) HandleTransfer(w http.ResponseWriter, r *http.Request, filePath, fileName string, fileSize int64) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	sessionID := fmt.Sprintf("%d", time.Now().UnixNano())
	session := &Session{
		ID:        sessionID,
		Conn:      conn,
		FileName:  fileName,
		FileSize:  fileSize,
		StartTime: time.Now(),
		Status:    "connected",
	}

	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.sessions, sessionID)
		s.mu.Unlock()
	}()

	handshake := Message{
		Type:     MsgHandshake,
		FileName: fileName,
		FileSize: fileSize,
	}
	if err := conn.WriteJSON(handshake); err != nil {
		logger.Error("发送握手消息失败: %v", err)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		s.sendError(conn, fmt.Sprintf("无法打开文件: %v", err))
		return
	}
	defer file.Close()

	buffer := make([]byte, 64*1024)
	var offset int64

	for {
		select {
		case <-r.Context().Done():
			return
		default:
		}

		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("读取消息失败: %v", err)
			}
			return
		}

		switch msg.Type {
		case MsgFileRequest:
			offset = msg.Offset
			if _, err := file.Seek(offset, io.SeekStart); err != nil {
				s.sendError(conn, fmt.Sprintf("文件定位失败: %v", err))
				return
			}

			session.Transferred = offset
			chunkSize := msg.ChunkLen
			if chunkSize <= 0 {
				chunkSize = 64 * 1024
			}

			for {
				n, err := file.Read(buffer)
				if n > 0 {
					dataMsg := Message{
						Type:   MsgFileData,
						Offset: offset,
						Data:   buffer[:n],
					}

					if err := conn.WriteJSON(dataMsg); err != nil {
						logger.Error("发送数据失败: %v", err)
						return
					}

					offset += int64(n)
					session.Transferred = offset

					stats.RecordSend(int64(n))
					stats.RecordDiskRead(int64(n))

					progress := float64(offset) / float64(fileSize) * 100
					progressMsg := Message{
						Type:     MsgProgress,
						Progress: progress,
					}
					conn.WriteJSON(progressMsg)
				}

				if err == io.EOF {
					endMsg := Message{Type: MsgFileEnd}
					conn.WriteJSON(endMsg)
					session.Status = "completed"
					logger.Info("WebSocket传输完成: %s", fileName)
					return
				}

				if err != nil {
					s.sendError(conn, fmt.Sprintf("读取文件失败: %v", err))
					return
				}
			}
		}
	}
}

func (s *Service) sendError(conn *websocket.Conn, errMsg string) {
	msg := Message{
		Type:  MsgError,
		Error: errMsg,
	}
	conn.WriteJSON(msg)
}

func (s *Service) GetSessions() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (s *Service) GetSessionsAsMap() []map[string]interface{} {
	sessions := s.GetSessions()
	result := make([]map[string]interface{}, len(sessions))
	for i, session := range sessions {
		result[i] = map[string]interface{}{
			"id":          session.ID,
			"file_name":   session.FileName,
			"file_size":   session.FileSize,
			"transferred": session.Transferred,
			"start_time":  session.StartTime.Format(time.RFC3339),
			"status":      session.Status,
		}
	}
	return result
}

type Client struct {
	conn       *websocket.Conn
	url        string
	progressCb func(progress float64)
}

func NewClient(url string) *Client {
	return &Client{url: url}
}

func (c *Client) SetProgressCallback(cb func(progress float64)) {
	c.progressCb = cb
}

func (c *Client) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) DownloadFile(savePath string) error {
	if c.conn == nil {
		return fmt.Errorf("未连接")
	}

	var handshake Message
	if err := c.conn.ReadJSON(&handshake); err != nil {
		return fmt.Errorf("读取握手消息失败: %w", err)
	}

	if handshake.Type != MsgHandshake {
		return fmt.Errorf("无效的握手消息")
	}

	fileName := handshake.FileName
	fileSize := handshake.FileSize

	destPath := filepath.Join(savePath, fileName)
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	reqMsg := Message{
		Type:     MsgFileRequest,
		Offset:   0,
		ChunkLen: 64 * 1024,
	}
	if err := c.conn.WriteJSON(reqMsg); err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}

	var received int64
	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			return fmt.Errorf("读取消息失败: %w", err)
		}

		switch msg.Type {
		case MsgFileData:
			if _, err := file.Write(msg.Data); err != nil {
				return fmt.Errorf("写入文件失败: %w", err)
			}
			received += int64(len(msg.Data))

			stats.RecordReceive(int64(len(msg.Data)))
			stats.RecordDiskWrite(int64(len(msg.Data)))

		case MsgProgress:
			if c.progressCb != nil {
				c.progressCb(msg.Progress)
			}

		case MsgFileEnd:
			logger.Info("WebSocket下载完成: %s (%d/%d bytes)", fileName, received, fileSize)
			return nil

		case MsgError:
			return fmt.Errorf("服务器错误: %s", msg.Error)
		}
	}
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
