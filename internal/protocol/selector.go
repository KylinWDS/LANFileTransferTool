package protocol

import (
	"fmt"
	"sync"
	"time"

	"lanfiletransfertool/internal/config"
	"lanfiletransfertool/pkg/logger"
)

// Protocol 传输协议类型
type Protocol string

const (
	ProtocolHTTP      Protocol = "http"
	ProtocolWebSocket Protocol = "websocket"
	ProtocolUDP       Protocol = "udp"
	ProtocolP2P       Protocol = "p2p"
	ProtocolAuto      Protocol = "auto"
)

// ProtocolInfo 协议信息
type ProtocolInfo struct {
	Type        Protocol
	Name        string
	Description string
	Priority    int  // 优先级，数字越小优先级越高
	Available   bool // 是否可用
	Latency     time.Duration
}

// Selector 协议选择器
type Selector struct {
	config    *config.Config
	protocols map[Protocol]*ProtocolInfo
	mu        sync.RWMutex
	preference Protocol // 用户偏好
}

// NewSelector 创建协议选择器
func NewSelector(cfg *config.Config) *Selector {
	s := &Selector{
		config:     cfg,
		protocols:  make(map[Protocol]*ProtocolInfo),
		preference: ProtocolAuto,
	}

	// 初始化协议信息
	s.protocols[ProtocolHTTP] = &ProtocolInfo{
		Type:        ProtocolHTTP,
		Name:        "HTTP",
		Description: "标准HTTP协议，兼容性最好",
		Priority:    4,
		Available:   true,
	}
	s.protocols[ProtocolWebSocket] = &ProtocolInfo{
		Type:        ProtocolWebSocket,
		Name:        "WebSocket",
		Description: "WebSocket协议，实时性好",
		Priority:    3,
		Available:   cfg.WebSocket.Enabled,
	}
	s.protocols[ProtocolUDP] = &ProtocolInfo{
		Type:        ProtocolUDP,
		Description: "UDP协议，速度快",
		Name:        "UDP",
		Priority:    2,
		Available:   cfg.UDP.Enabled,
	}
	s.protocols[ProtocolP2P] = &ProtocolInfo{
		Type:        ProtocolP2P,
		Name:        "P2P",
		Description: "P2P直连，无需服务器中转",
		Priority:    1,
		Available:   cfg.P2P.Enabled,
	}

	return s
}

// SelectionCriteria 协议选择条件
type SelectionCriteria struct {
	FileSize       int64
	PeerAvailable  bool   // 是否支持P2P直连
	NetworkType    string // "lan" | "wan"
	RequireRealtime bool  // 是否需要实时性
	UserOverride   Protocol // 用户手动覆盖
}

// Select 根据条件选择最佳协议
func (s *Selector) Select(criteria SelectionCriteria) Protocol {
	// 如果用户手动指定了协议，优先使用
	if criteria.UserOverride != "" && criteria.UserOverride != ProtocolAuto {
		if s.isAvailable(criteria.UserOverride) {
			logger.Info("使用用户指定的协议: %s", criteria.UserOverride)
			return criteria.UserOverride
		}
		logger.Warn("用户指定的协议 %s 不可用，将自动选择", criteria.UserOverride)
	}

	// 如果用户设置了偏好协议（非auto），优先尝试
	if s.preference != "" && s.preference != ProtocolAuto {
		if s.isAvailable(s.preference) {
			// 检查偏好协议是否适合当前场景
			if s.isSuitable(s.preference, criteria) {
				logger.Info("使用偏好协议: %s", s.preference)
				return s.preference
			}
		}
	}

	// 智能选择逻辑
	selected := s.smartSelect(criteria)
	logger.Info("智能选择协议: %s (文件大小: %d, P2P可用: %v)", 
		selected, criteria.FileSize, criteria.PeerAvailable)
	
	return selected
}

// smartSelect 智能选择算法
func (s *Selector) smartSelect(criteria SelectionCriteria) Protocol {
	// 场景1：大文件 (>100MB) - 优先P2P或UDP
	if criteria.FileSize > 100*1024*1024 {
		if criteria.PeerAvailable && s.isAvailable(ProtocolP2P) {
			return ProtocolP2P
		}
		if s.isAvailable(ProtocolUDP) {
			return ProtocolUDP
		}
	}

	// 场景2：中等文件 (10MB-100MB) - 优先WebSocket或UDP
	if criteria.FileSize > 10*1024*1024 {
		if criteria.RequireRealtime && s.isAvailable(ProtocolWebSocket) {
			return ProtocolWebSocket
		}
		if s.isAvailable(ProtocolUDP) {
			return ProtocolUDP
		}
		if s.isAvailable(ProtocolWebSocket) {
			return ProtocolWebSocket
		}
	}

	// 场景3：需要实时性 - WebSocket
	if criteria.RequireRealtime && s.isAvailable(ProtocolWebSocket) {
		return ProtocolWebSocket
	}

	// 场景4：局域网环境 - 可以尝试P2P
	if criteria.NetworkType == "lan" && criteria.PeerAvailable && s.isAvailable(ProtocolP2P) {
		return ProtocolP2P
	}

	// 场景5：小文件或默认情况 - HTTP（最稳定）
	if s.isAvailable(ProtocolHTTP) {
		return ProtocolHTTP
	}

	// 降级选择：按优先级选择任意可用协议
	for _, proto := range []Protocol{ProtocolWebSocket, ProtocolUDP, ProtocolP2P} {
		if s.isAvailable(proto) {
			return proto
		}
	}

	// 最终 fallback
	return ProtocolHTTP
}

// isAvailable 检查协议是否可用
func (s *Selector) isAvailable(proto Protocol) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.protocols[proto]
	if !exists {
		return false
	}
	return info.Available
}

// isSuitable 检查协议是否适合当前场景
func (s *Selector) isSuitable(proto Protocol, criteria SelectionCriteria) bool {
	switch proto {
	case ProtocolP2P:
		// P2P需要对方支持
		return criteria.PeerAvailable
	case ProtocolUDP:
		// UDP适合大文件，但不适合需要高可靠性的场景
		return criteria.FileSize > 1024*1024 // 大于1MB
	case ProtocolWebSocket:
		// WebSocket适合需要实时性的场景
		return true
	case ProtocolHTTP:
		// HTTP通用
		return true
	}
	return true
}

// SetPreference 设置用户偏好协议
func (s *Selector) SetPreference(proto Protocol) error {
	if proto != ProtocolAuto && proto != ProtocolHTTP && proto != ProtocolWebSocket && 
		proto != ProtocolUDP && proto != ProtocolP2P {
		return fmt.Errorf("无效的协议类型: %s", proto)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.preference = proto
	logger.Info("设置协议偏好: %s", proto)
	return nil
}

// GetPreference 获取用户偏好协议
func (s *Selector) GetPreference() Protocol {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.preference
}

// UpdateProtocolStatus 更新协议可用状态
func (s *Selector) UpdateProtocolStatus(proto Protocol, available bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if info, exists := s.protocols[proto]; exists {
		info.Available = available
		logger.Info("更新协议状态: %s = %v", proto, available)
	}
}

// GetAllProtocols 获取所有协议信息
func (s *Selector) GetAllProtocols() []ProtocolInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]ProtocolInfo, 0, len(s.protocols))
	for _, info := range s.protocols {
		result = append(result, *info)
	}
	return result
}

// GetProtocolInfo 获取指定协议信息
func (s *Selector) GetProtocolInfo(proto Protocol) (ProtocolInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.protocols[proto]
	if !exists {
		return ProtocolInfo{}, false
	}
	return *info, true
}

// RecommendProtocol 根据文件大小推荐协议（用于UI显示）
func (s *Selector) RecommendProtocol(fileSize int64) Protocol {
	if fileSize > 100*1024*1024 {
		return ProtocolP2P
	}
	if fileSize > 10*1024*1024 {
		return ProtocolUDP
	}
	if fileSize > 1024*1024 {
		return ProtocolWebSocket
	}
	return ProtocolHTTP
}
