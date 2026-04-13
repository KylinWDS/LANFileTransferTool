package udp

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"lanfiletransfertool/internal/stats"
	"lanfiletransfertool/pkg/logger"
)

const (
	MaxPacketSize    = 65507
	HeaderSize       = 32
	HandshakeMsg     = "UDP_TRANSFER_HANDSHAKE"
	HandshakeAck     = "UDP_TRANSFER_ACK"
	FileRequestMsg   = "FILE_REQUEST"
	FileDataMsg      = "FILE_DATA"
	FileEndMsg       = "FILE_END"
	ErrorMsg         = "ERROR"
	ProgressMsg      = "PROGRESS"
	ResendRequestMsg = "RESEND"
	DefaultTimeout   = 30 * time.Second
	MaxRetries       = 3
)

type PacketHeader struct {
	Magic      uint32
	Type       uint8
	Flags      uint8
	Reserved   uint16
	Sequence   uint32
	TotalSize  uint64
	ChunkIndex uint32
	ChunkCount uint32
	DataSize   uint32
}

type TransferSession struct {
	ID           string
	RemoteAddr   *net.UDPAddr
	FileName     string
	FileSize     int64
	ReceivedMap  map[uint32]bool
	Chunks       [][]byte
	StartTime    time.Time
	LastActivity time.Time
	Status       string
}

type Service struct {
	port       int
	chunkSize  int
	conn       *net.UDPConn
	sessions   map[string]*TransferSession
	mu         sync.RWMutex
	running    bool
	stopChan   chan struct{}
	fileGetter func(fileID string) (string, string, int64, error)
}

// UDPConfig UDP配置接口
type UDPConfig interface {
	GetUDPPort() int
}

func NewService(cfg UDPConfig, fileGetter func(fileID string) (string, string, int64, error)) *Service {
	port := cfg.GetUDPPort()
	if port <= 0 {
		port = 37022 // 默认端口
	}
	return &Service{
		port:       port,
		chunkSize:  32 * 1024, // 默认分块大小 32KB
		sessions:   make(map[string]*TransferSession),
		stopChan:   make(chan struct{}),
		fileGetter: fileGetter,
	}
}

func (s *Service) Start() error {
	addr := &net.UDPAddr{
		Port: s.port,
		IP:   net.IPv4zero,
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return fmt.Errorf("监听UDP端口失败: %w", err)
	}

	s.conn = conn
	s.running = true

	go s.listen()

	logger.Info("UDP传输服务已启动，端口: %d", s.port)
	return nil
}

func (s *Service) Stop() {
	if !s.running {
		return
	}

	s.running = false
	close(s.stopChan)

	if s.conn != nil {
		s.conn.Close()
	}

	logger.Info("UDP传输服务已停止")
}

func (s *Service) listen() {
	buffer := make([]byte, MaxPacketSize)

	for s.running {
		n, remoteAddr, err := s.conn.ReadFromUDP(buffer)
		if err != nil {
			if s.running {
				logger.Warn("读取UDP数据失败: %v", err)
			}
			continue
		}

		if n < HeaderSize {
			continue
		}

		packet := make([]byte, n)
		copy(packet, buffer[:n])

		go s.handlePacket(remoteAddr, packet)
	}
}

func (s *Service) handlePacket(remoteAddr *net.UDPAddr, packet []byte) {
	header := s.parseHeader(packet[:HeaderSize])
	if header.Magic != 0x4C465458 {
		return
	}

	switch header.Type {
	case 0x01:
		s.handleHandshake(remoteAddr, packet[HeaderSize:])
	case 0x02:
		s.handleFileRequest(remoteAddr, packet[HeaderSize:])
	case 0x03:
		s.handleFileData(remoteAddr, header, packet[HeaderSize:])
	case 0x04:
		s.handleFileEnd(remoteAddr, packet[HeaderSize:])
	case 0x05:
		s.handleResendRequest(remoteAddr, packet[HeaderSize:])
	}
}

func (s *Service) parseHeader(data []byte) *PacketHeader {
	return &PacketHeader{
		Magic:      binary.BigEndian.Uint32(data[0:4]),
		Type:       data[4],
		Flags:      data[5],
		Reserved:   binary.BigEndian.Uint16(data[6:8]),
		Sequence:   binary.BigEndian.Uint32(data[8:12]),
		TotalSize:  binary.BigEndian.Uint64(data[12:20]),
		ChunkIndex: binary.BigEndian.Uint32(data[20:24]),
		ChunkCount: binary.BigEndian.Uint32(data[24:28]),
		DataSize:   binary.BigEndian.Uint32(data[28:32]),
	}
}

func (s *Service) buildHeader(header *PacketHeader) []byte {
	data := make([]byte, HeaderSize)
	binary.BigEndian.PutUint32(data[0:4], header.Magic)
	data[4] = header.Type
	data[5] = header.Flags
	binary.BigEndian.PutUint16(data[6:8], header.Reserved)
	binary.BigEndian.PutUint32(data[8:12], header.Sequence)
	binary.BigEndian.PutUint64(data[12:20], header.TotalSize)
	binary.BigEndian.PutUint32(data[20:24], header.ChunkIndex)
	binary.BigEndian.PutUint32(data[24:28], header.ChunkCount)
	binary.BigEndian.PutUint32(data[28:32], header.DataSize)
	return data
}

func (s *Service) handleHandshake(remoteAddr *net.UDPAddr, data []byte) {
	var msg struct {
		FileID string `json:"file_id"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	filePath, fileName, fileSize, err := s.fileGetter(msg.FileID)
	if err != nil {
		s.sendError(remoteAddr, "文件不存在")
		return
	}

	sessionID := fmt.Sprintf("%d", time.Now().UnixNano())
	session := &TransferSession{
		ID:           sessionID,
		RemoteAddr:   remoteAddr,
		FileName:     fileName,
		FileSize:     fileSize,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		Status:       "handshake",
	}

	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()

	ack := struct {
		SessionID string `json:"session_id"`
		FileName  string `json:"file_name"`
		FileSize  int64  `json:"file_size"`
	}{
		SessionID: sessionID,
		FileName:  fileName,
		FileSize:  fileSize,
	}

	ackData, _ := json.Marshal(ack)
	header := &PacketHeader{
		Magic: 0x4C465458,
		Type:  0x01,
	}
	packet := append(s.buildHeader(header), ackData...)
	s.conn.WriteToUDP(packet, remoteAddr)

	go s.sendFile(sessionID, filePath)
}

func (s *Service) sendFile(sessionID, filePath string) {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		s.sendError(session.RemoteAddr, "无法打开文件")
		return
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	chunkCount := uint32((fileSize + int64(s.chunkSize) - 1) / int64(s.chunkSize))

	buffer := make([]byte, s.chunkSize)
	var chunkIndex uint32 = 0

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			header := &PacketHeader{
				Magic:      0x4C465458,
				Type:       0x03,
				Sequence:   uint32(time.Now().UnixNano()),
				TotalSize:  uint64(fileSize),
				ChunkIndex: chunkIndex,
				ChunkCount: chunkCount,
				DataSize:   uint32(n),
			}

			packet := append(s.buildHeader(header), buffer[:n]...)
			s.conn.WriteToUDP(packet, session.RemoteAddr)

			stats.RecordSend(int64(n))
			stats.RecordDiskRead(int64(n))

			chunkIndex++
		}

		if err == io.EOF {
			header := &PacketHeader{
				Magic: 0x4C465458,
				Type:  0x04,
			}
			packet := append(s.buildHeader(header), []byte(sessionID)...)
			s.conn.WriteToUDP(packet, session.RemoteAddr)
			break
		}

		if err != nil {
			break
		}
	}

	s.mu.Lock()
	session.Status = "completed"
	s.mu.Unlock()
}

func (s *Service) handleFileRequest(remoteAddr *net.UDPAddr, data []byte) {
}

func (s *Service) handleFileData(remoteAddr *net.UDPAddr, header *PacketHeader, data []byte) {
}

func (s *Service) handleFileEnd(remoteAddr *net.UDPAddr, data []byte) {
}

func (s *Service) handleResendRequest(remoteAddr *net.UDPAddr, data []byte) {
}

func (s *Service) sendError(remoteAddr *net.UDPAddr, errMsg string) {
	header := &PacketHeader{
		Magic: 0x4C465458,
		Type:  0xFF,
	}
	packet := append(s.buildHeader(header), []byte(errMsg)...)
	s.conn.WriteToUDP(packet, remoteAddr)
}

func (s *Service) GetSessions() []*TransferSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*TransferSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

type Client struct {
	conn     *net.UDPConn
	server   *net.UDPAddr
	timeout  time.Duration
	progress func(progress float64)
}

func NewClient(serverAddr string, port int) (*Client, error) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", serverAddr, port))
	if err != nil {
		return nil, err
	}

	localAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	conn, err := net.DialUDP("udp4", localAddr, addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:    conn,
		server:  addr,
		timeout: DefaultTimeout,
	}, nil
}

func (c *Client) SetProgressCallback(cb func(progress float64)) {
	c.progress = cb
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) DownloadFile(fileID, savePath string) error {
	handshake := struct {
		FileID string `json:"file_id"`
	}{
		FileID: fileID,
	}

	handshakeData, _ := json.Marshal(handshake)
	header := &PacketHeader{
		Magic: 0x4C465458,
		Type:  0x01,
	}
	packet := append(c.buildHeader(header), handshakeData...)

	if _, err := c.conn.Write(packet); err != nil {
		return err
	}

	buffer := make([]byte, MaxPacketSize)
	c.conn.SetReadDeadline(time.Now().Add(c.timeout))

	n, err := c.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("等待握手响应超时")
	}

	ackHeader := c.parseHeader(buffer[:HeaderSize])
	if ackHeader.Type != 0x01 {
		return fmt.Errorf("无效的握手响应")
	}

	var ack struct {
		SessionID string `json:"session_id"`
		FileName  string `json:"file_name"`
		FileSize  int64  `json:"file_size"`
	}
	if err := json.Unmarshal(buffer[HeaderSize:n], &ack); err != nil {
		return err
	}

	destPath := filepath.Join(savePath, ack.FileName)
	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	chunks := make(map[uint32][]byte)
	chunkCount := uint32((ack.FileSize + int64(32*1024) - 1) / int64(32*1024))

	for {
		c.conn.SetReadDeadline(time.Now().Add(c.timeout))
		n, err := c.conn.Read(buffer)
		if err != nil {
			break
		}

		header := c.parseHeader(buffer[:HeaderSize])

		switch header.Type {
		case 0x03:
			data := make([]byte, header.DataSize)
			copy(data, buffer[HeaderSize:HeaderSize+header.DataSize])
			chunks[header.ChunkIndex] = data

			stats.RecordReceive(int64(header.DataSize))
			stats.RecordDiskWrite(int64(header.DataSize))

			if c.progress != nil {
				progress := float64(len(chunks)) / float64(chunkCount) * 100
				c.progress(progress)
			}

		case 0x04:
			for i := uint32(0); i < chunkCount; i++ {
				if chunk, ok := chunks[i]; ok {
					file.Write(chunk)
				}
			}
			logger.Info("UDP下载完成: %s", ack.FileName)
			return nil

		case 0xFF:
			return fmt.Errorf("服务器错误: %s", string(buffer[HeaderSize:n]))
		}
	}

	return nil
}

func (c *Client) parseHeader(data []byte) *PacketHeader {
	return &PacketHeader{
		Magic:      binary.BigEndian.Uint32(data[0:4]),
		Type:       data[4],
		Flags:      data[5],
		Reserved:   binary.BigEndian.Uint16(data[6:8]),
		Sequence:   binary.BigEndian.Uint32(data[8:12]),
		TotalSize:  binary.BigEndian.Uint64(data[12:20]),
		ChunkIndex: binary.BigEndian.Uint32(data[20:24]),
		ChunkCount: binary.BigEndian.Uint32(data[24:28]),
		DataSize:   binary.BigEndian.Uint32(data[28:32]),
	}
}

func (c *Client) buildHeader(header *PacketHeader) []byte {
	data := make([]byte, HeaderSize)
	binary.BigEndian.PutUint32(data[0:4], header.Magic)
	data[4] = header.Type
	data[5] = header.Flags
	binary.BigEndian.PutUint16(data[6:8], header.Reserved)
	binary.BigEndian.PutUint32(data[8:12], header.Sequence)
	binary.BigEndian.PutUint64(data[12:20], header.TotalSize)
	binary.BigEndian.PutUint32(data[20:24], header.ChunkIndex)
	binary.BigEndian.PutUint32(data[24:28], header.ChunkCount)
	binary.BigEndian.PutUint32(data[28:32], header.DataSize)
	return data
}
