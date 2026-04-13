package discovery

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"lanfiletransfertool/pkg/logger"
)

const (
	DiscoveryPort     = 37021
	MsgTypeDiscovery  = "LANFILETRANSFER_DISCOVERY"
	MsgTypeResponse   = "LANFILETRANSFER_RESPONSE"
	BroadcastInterval = 5 * time.Second
	PeerTimeout       = 30 * time.Second
)

type Peer struct {
	IP       string    `json:"ip"`
	Port     int       `json:"port"`
	Name     string    `json:"name"`
	LastSeen time.Time `json:"last_seen"`
}

type DiscoveryMessage struct {
	Type    string `json:"type"`
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Service struct {
	port       int
	serverPort int
	name       string
	version    string
	conn       *net.UDPConn
	peers      map[string]*Peer
	mu         sync.RWMutex
	running    bool
	stopChan   chan struct{}
	localIPs   []string
}

func NewService(serverPort int, name string) *Service {
	return &Service{
		port:       DiscoveryPort,
		serverPort: serverPort,
		name:       name,
		version:    "0.2.0",
		peers:      make(map[string]*Peer),
		stopChan:   make(chan struct{}),
		localIPs:   getLocalIPs(),
	}
}

func getLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips
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
	go s.broadcast()

	logger.Info("UDP发现服务已启动，端口: %d", s.port)
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

	logger.Info("UDP发现服务已停止")
}

func (s *Service) listen() {
	buf := make([]byte, 1024)

	for s.running {
		n, remoteAddr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			if s.running {
				logger.Warn("读取UDP消息失败: %v", err)
			}
			continue
		}

		var msg DiscoveryMessage
		if err := json.Unmarshal(buf[:n], &msg); err != nil {
			continue
		}

		if s.isLocalIP(remoteAddr.IP.String()) {
			continue
		}

		switch msg.Type {
		case MsgTypeDiscovery:
			s.handleDiscovery(remoteAddr.IP.String(), &msg)
		case MsgTypeResponse:
			s.handleResponse(remoteAddr.IP.String(), &msg)
		}
	}
}

func (s *Service) isLocalIP(ip string) bool {
	for _, localIP := range s.localIPs {
		if ip == localIP {
			return true
		}
	}
	return false
}

func (s *Service) handleDiscovery(remoteIP string, msg *DiscoveryMessage) {
	response := DiscoveryMessage{
		Type:    MsgTypeResponse,
		IP:      s.getPreferredIP(),
		Port:    s.serverPort,
		Name:    s.name,
		Version: s.version,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return
	}

	remoteAddr := &net.UDPAddr{
		IP:   net.ParseIP(remoteIP),
		Port: s.port,
	}

	s.conn.WriteToUDP(data, remoteAddr)
}

func (s *Service) handleResponse(remoteIP string, msg *DiscoveryMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peer := &Peer{
		IP:       msg.IP,
		Port:     msg.Port,
		Name:     msg.Name,
		LastSeen: time.Now(),
	}

	key := fmt.Sprintf("%s:%d", peer.IP, peer.Port)
	s.peers[key] = peer

	logger.Debug("发现设备: %s (%s:%d)", peer.Name, peer.IP, peer.Port)
}

func (s *Service) broadcast() {
	ticker := time.NewTicker(BroadcastInterval)
	defer ticker.Stop()

	s.sendDiscovery()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.sendDiscovery()
			s.cleanupPeers()
		}
	}
}

func (s *Service) sendDiscovery() {
	msg := DiscoveryMessage{
		Type:    MsgTypeDiscovery,
		IP:      s.getPreferredIP(),
		Port:    s.serverPort,
		Name:    s.name,
		Version: s.version,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	broadcastAddr := &net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: s.port,
	}

	s.conn.WriteToUDP(data, broadcastAddr)
}

func (s *Service) getPreferredIP() string {
	if len(s.localIPs) > 0 {
		for _, ip := range s.localIPs {
			if ip[0:3] == "192" || ip[0:3] == "10." || (ip[0:3] == "172" && len(ip) > 6) {
				return ip
			}
		}
		return s.localIPs[0]
	}
	return "127.0.0.1"
}

func (s *Service) cleanupPeers() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, peer := range s.peers {
		if now.Sub(peer.LastSeen) > PeerTimeout {
			delete(s.peers, key)
		}
	}
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

func (s *Service) GetPeersAsMap() []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(s.peers))
	for _, peer := range s.peers {
		result = append(result, map[string]interface{}{
			"ip":        peer.IP,
			"port":      peer.Port,
			"name":      peer.Name,
			"last_seen": peer.LastSeen.Format(time.RFC3339),
		})
	}
	return result
}
