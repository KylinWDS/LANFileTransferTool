package p2p

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"lanfiletransfertool/internal/stats"
	"lanfiletransfertool/pkg/logger"
)

const (
	DefaultP2PPort    = 37023
	HandshakeTimeout  = 10 * time.Second
	TransferTimeout   = 300 * time.Second
	ChunkSize         = 64 * 1024
	MaxPendingChunks  = 16
)

type MessageType int

const (
	MsgHandshake MessageType = iota
	MsgHandshakeAck
	MsgFileRequest
	MsgFileInfo
	MsgChunkRequest
	MsgChunkData
	MsgChunkAck
	MsgFileEnd
	MsgError
	MsgCancel
)

type Message struct {
	Type       MessageType `json:"type"`
	SenderID   string      `json:"sender_id,omitempty"`
	ReceiverID string      `json:"receiver_id,omitempty"`
	FileID     string      `json:"file_id,omitempty"`
	FileName   string      `json:"file_name,omitempty"`
	FileSize   int64       `json:"file_size,omitempty"`
	ChunkIndex int         `json:"chunk_index,omitempty"`
	ChunkCount int         `json:"chunk_count,omitempty"`
	ChunkSize  int         `json:"chunk_size,omitempty"`
	Data       []byte      `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
}

type Peer struct {
	ID        string
	IP        string
	Port      int
	Conn      net.Conn
	LastSeen  time.Time
	IsActive  bool
}

type Transfer struct {
	ID           string
	Peer         *Peer
	FileName     string
	FileSize     int64
	FilePath     string
	Chunks       map[int][]byte
	ReceivedMap  map[int]bool
	ChunkCount   int
	Progress     float64
	Status       string
	StartTime    time.Time
	LastActivity time.Time
}

type Service struct {
	port       int
	peerID     string
	listener   net.Listener
	peers      map[string]*Peer
	transfers  map[string]*Transfer
	mu         sync.RWMutex
	running    bool
	stopChan   chan struct{}
	fileGetter func(fileID string) (string, string, int64, error)
	onProgress func(transferID string, progress float64)
}

func NewService(port int, fileGetter func(fileID string) (string, string, int64, error)) *Service {
	if port <= 0 {
		port = DefaultP2PPort
	}
	return &Service{
		port:       port,
		peerID:     generatePeerID(),
		peers:      make(map[string]*Peer),
		transfers:  make(map[string]*Transfer),
		stopChan:   make(chan struct{}),
		fileGetter: fileGetter,
	}
}

func generatePeerID() string {
	return fmt.Sprintf("P2P-%d", time.Now().UnixNano())
}

func (s *Service) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("监听P2P端口失败: %w", err)
	}

	s.listener = listener
	s.running = true

	go s.acceptConnections()

	logger.Info("P2P传输服务已启动，端口: %d", s.port)
	return nil
}

func (s *Service) Stop() {
	if !s.running {
		return
	}

	s.running = false
	close(s.stopChan)

	if s.listener != nil {
		s.listener.Close()
	}

	s.mu.Lock()
	for _, peer := range s.peers {
		if peer.Conn != nil {
			peer.Conn.Close()
		}
	}
	for _, transfer := range s.transfers {
		transfer.Status = "cancelled"
	}
	s.mu.Unlock()

	logger.Info("P2P传输服务已停止")
}

func (s *Service) acceptConnections() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				logger.Warn("接受P2P连接失败: %v", err)
			}
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Service) handleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	logger.Info("P2P连接: %s", remoteAddr)

	conn.SetDeadline(time.Now().Add(HandshakeTimeout))

	var handshake Message
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&handshake); err != nil {
		logger.Warn("P2P握手失败: %v", err)
		return
	}

	if handshake.Type != MsgHandshake {
		return
	}

	peer := &Peer{
		ID:       handshake.SenderID,
		Conn:     conn,
		LastSeen: time.Now(),
		IsActive: true,
	}

	if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		peer.IP = tcpAddr.IP.String()
		peer.Port = tcpAddr.Port
	}

	s.mu.Lock()
	s.peers[peer.ID] = peer
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.peers, peer.ID)
		s.mu.Unlock()
	}()

	ack := Message{
		Type:     MsgHandshakeAck,
		SenderID: s.peerID,
	}
	encoder := json.NewEncoder(conn)
	encoder.Encode(ack)

	conn.SetDeadline(time.Time{})

	s.handleMessages(peer, decoder, encoder)
}

func (s *Service) handleMessages(peer *Peer, decoder *json.Decoder, encoder *json.Encoder) {
	for s.running {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			if s.running {
				logger.Debug("P2P连接关闭: %s", peer.ID)
			}
			break
		}

		peer.LastSeen = time.Now()

		switch msg.Type {
		case MsgFileRequest:
			s.handleFileRequest(peer, msg, encoder)
		case MsgChunkRequest:
			s.handleChunkRequest(peer, msg, encoder)
		case MsgChunkData:
			s.handleChunkData(peer, msg, encoder)
		case MsgChunkAck:
			s.handleChunkAck(peer, msg)
		case MsgFileEnd:
			s.handleFileEnd(peer, msg)
		case MsgCancel:
			s.handleCancel(peer, msg)
		}
	}
}

func (s *Service) handleFileRequest(peer *Peer, msg Message, encoder *json.Encoder) {
	filePath, fileName, fileSize, err := s.fileGetter(msg.FileID)
	if err != nil {
		encoder.Encode(Message{Type: MsgError, Error: "文件不存在"})
		return
	}

	transferID := generateTransferID()
	chunkCount := int((fileSize + ChunkSize - 1) / ChunkSize)

	transfer := &Transfer{
		ID:           transferID,
		Peer:         peer,
		FileName:     fileName,
		FileSize:     fileSize,
		FilePath:     filePath,
		ChunkCount:   chunkCount,
		ReceivedMap:  make(map[int]bool),
		Status:       "sending",
		StartTime:    time.Now(),
		LastActivity: time.Now(),
	}

	s.mu.Lock()
	s.transfers[transferID] = transfer
	s.mu.Unlock()

	info := Message{
		Type:       MsgFileInfo,
		FileID:     transferID,
		FileName:   fileName,
		FileSize:   fileSize,
		ChunkCount: chunkCount,
		ChunkSize:  ChunkSize,
	}
	encoder.Encode(info)
}

func (s *Service) handleChunkRequest(peer *Peer, msg Message, encoder *json.Encoder) {
	s.mu.RLock()
	transfer, exists := s.transfers[msg.FileID]
	s.mu.RUnlock()

	if !exists {
		encoder.Encode(Message{Type: MsgError, Error: "传输不存在"})
		return
	}

	file, err := os.Open(transfer.FilePath)
	if err != nil {
		encoder.Encode(Message{Type: MsgError, Error: "无法打开文件"})
		return
	}
	defer file.Close()

	offset := int64(msg.ChunkIndex * ChunkSize)
	file.Seek(offset, io.SeekStart)

	buffer := make([]byte, ChunkSize)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		encoder.Encode(Message{Type: MsgError, Error: "读取文件失败"})
		return
	}

	chunkData := buffer[:n]
	dataMsg := Message{
		Type:       MsgChunkData,
		FileID:     transfer.ID,
		ChunkIndex: msg.ChunkIndex,
		Data:       chunkData,
	}

	if err := encoder.Encode(&dataMsg); err != nil {
		logger.Error("发送块数据失败: %v", err)
		return
	}

	stats.RecordSend(int64(n))
	stats.RecordDiskRead(int64(n))
}

func (s *Service) handleChunkData(peer *Peer, msg Message, encoder *json.Encoder) {
	s.mu.Lock()
	transfer, exists := s.transfers[msg.FileID]
	if exists {
		if transfer.Chunks == nil {
			transfer.Chunks = make(map[int][]byte)
		}
		transfer.Chunks[msg.ChunkIndex] = msg.Data
		transfer.ReceivedMap[msg.ChunkIndex] = true
		transfer.LastActivity = time.Now()

		progress := float64(len(transfer.ReceivedMap)) / float64(transfer.ChunkCount) * 100
		transfer.Progress = progress

		if s.onProgress != nil {
			s.onProgress(msg.FileID, progress)
		}

		stats.RecordReceive(int64(len(msg.Data)))
		stats.RecordDiskWrite(int64(len(msg.Data)))
	}
	s.mu.Unlock()

	ack := Message{
		Type:       MsgChunkAck,
		FileID:     msg.FileID,
		ChunkIndex: msg.ChunkIndex,
	}
	encoder.Encode(ack)

	if exists && len(transfer.ReceivedMap) == transfer.ChunkCount {
		s.completeTransfer(msg.FileID)
	}
}

func (s *Service) handleChunkAck(peer *Peer, msg Message) {
	s.mu.Lock()
	if transfer, exists := s.transfers[msg.FileID]; exists {
		transfer.LastActivity = time.Now()
	}
	s.mu.Unlock()
}

func (s *Service) handleFileEnd(peer *Peer, msg Message) {
	s.mu.Lock()
	if transfer, exists := s.transfers[msg.FileID]; exists {
		transfer.Status = "completed"
		transfer.Progress = 100
	}
	s.mu.Unlock()
}

func (s *Service) handleCancel(peer *Peer, msg Message) {
	s.mu.Lock()
	if transfer, exists := s.transfers[msg.FileID]; exists {
		transfer.Status = "cancelled"
	}
	s.mu.Unlock()
}

func (s *Service) completeTransfer(transferID string) {
	s.mu.Lock()
	transfer, exists := s.transfers[transferID]
	if !exists {
		s.mu.Unlock()
		return
	}

	file, err := os.Create(transfer.FilePath)
	if err != nil {
		transfer.Status = "failed"
		s.mu.Unlock()
		return
	}
	defer file.Close()

	for i := 0; i < transfer.ChunkCount; i++ {
		if chunk, ok := transfer.Chunks[i]; ok {
			file.Write(chunk)
		}
	}

	transfer.Status = "completed"
	transfer.Progress = 100
	s.mu.Unlock()

	logger.Info("P2P传输完成: %s", transfer.FileName)
}

func (s *Service) ConnectToPeer(ip string, port int) (*Peer, error) {
	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", addr, HandshakeTimeout)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}

	handshake := Message{
		Type:     MsgHandshake,
		SenderID: s.peerID,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(&handshake); err != nil {
		conn.Close()
		return nil, err
	}

	conn.SetDeadline(time.Now().Add(HandshakeTimeout))

	var ack Message
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&ack); err != nil {
		conn.Close()
		return nil, err
	}

	if ack.Type != MsgHandshakeAck {
		conn.Close()
		return nil, fmt.Errorf("无效的握手响应")
	}

	conn.SetDeadline(time.Time{})

	peer := &Peer{
		ID:       ack.SenderID,
		IP:       ip,
		Port:     port,
		Conn:     conn,
		LastSeen: time.Now(),
		IsActive: true,
	}

	s.mu.Lock()
	s.peers[peer.ID] = peer
	s.mu.Unlock()

	go s.handleMessages(peer, decoder, encoder)

	return peer, nil
}

func (s *Service) RequestFile(peerID, fileID, savePath string, progressCb func(progress float64)) error {
	s.mu.RLock()
	peer, exists := s.peers[peerID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("对等节点不存在")
	}

	encoder := json.NewEncoder(peer.Conn)

	req := Message{
		Type:   MsgFileRequest,
		FileID: fileID,
	}
	if err := encoder.Encode(&req); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetPeers() []*Peer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	peers := make([]*Peer, 0, len(s.peers))
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	return peers
}

func (s *Service) GetTransfers() []*Transfer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	transfers := make([]*Transfer, 0, len(s.transfers))
	for _, transfer := range s.transfers {
		transfers = append(transfers, transfer)
	}
	return transfers
}

func (s *Service) SetProgressCallback(cb func(transferID string, progress float64)) {
	s.onProgress = cb
}

func generateTransferID() string {
	return fmt.Sprintf("TRF-%d", time.Now().UnixNano())
}

func (s *Service) GetPeerID() string {
	return s.peerID
}

func (s *Service) GetPort() int {
	return s.port
}

func (t *Transfer) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":           t.ID,
		"file_name":    t.FileName,
		"file_size":    t.FileSize,
		"progress":     t.Progress,
		"status":       t.Status,
		"start_time":   t.StartTime.Format(time.RFC3339),
		"chunk_count":  t.ChunkCount,
		"received":     len(t.ReceivedMap),
	}
}

func (p *Peer) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":        p.ID,
		"ip":        p.IP,
		"port":      p.Port,
		"last_seen": p.LastSeen.Format(time.RFC3339),
		"is_active": p.IsActive,
	}
}
