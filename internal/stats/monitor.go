package stats

import (
	"sync"
	"time"
)

type TransferStats struct {
	BytesSent     int64
	BytesReceived int64
	StartTime     time.Time
}

type SpeedSample struct {
	Timestamp time.Time
	Bytes     int64
}

type Monitor struct {
	sendHistory     []SpeedSample
	receiveHistory  []SpeedSample
	diskReadHistory []SpeedSample
	diskWriteHistory []SpeedSample
	mu              sync.RWMutex
	maxSamples      int
	windowSize      time.Duration
}

var globalMonitor *Monitor
var once sync.Once

func GetMonitor() *Monitor {
	once.Do(func() {
		globalMonitor = &Monitor{
			maxSamples: 60,
			windowSize: 60 * time.Second,
		}
	})
	return globalMonitor
}

func NewMonitor() *Monitor {
	return &Monitor{
		maxSamples: 60,
		windowSize: 60 * time.Second,
	}
}

func (m *Monitor) RecordSend(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	m.sendHistory = append(m.sendHistory, SpeedSample{
		Timestamp: now,
		Bytes:     bytes,
	})

	m.cleanupHistory(&m.sendHistory)
}

func (m *Monitor) RecordReceive(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	m.receiveHistory = append(m.receiveHistory, SpeedSample{
		Timestamp: now,
		Bytes:     bytes,
	})

	m.cleanupHistory(&m.receiveHistory)
}

func (m *Monitor) RecordDiskRead(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	m.diskReadHistory = append(m.diskReadHistory, SpeedSample{
		Timestamp: now,
		Bytes:     bytes,
	})

	m.cleanupHistory(&m.diskReadHistory)
}

func (m *Monitor) RecordDiskWrite(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	m.diskWriteHistory = append(m.diskWriteHistory, SpeedSample{
		Timestamp: now,
		Bytes:     bytes,
	})

	m.cleanupHistory(&m.diskWriteHistory)
}

func (m *Monitor) cleanupHistory(history *[]SpeedSample) {
	now := time.Now()
	cutoff := now.Add(-m.windowSize)

	validIdx := 0
	for _, sample := range *history {
		if sample.Timestamp.After(cutoff) {
			(*history)[validIdx] = sample
			validIdx++
		}
	}
	*history = (*history)[:validIdx]

	if len(*history) > m.maxSamples {
		*history = (*history)[len(*history)-m.maxSamples:]
	}
}

func (m *Monitor) GetSendSpeed() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.calculateSpeed(m.sendHistory)
}

func (m *Monitor) GetReceiveSpeed() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.calculateSpeed(m.receiveHistory)
}

func (m *Monitor) GetDiskReadSpeed() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.calculateSpeed(m.diskReadHistory)
}

func (m *Monitor) GetDiskWriteSpeed() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.calculateSpeed(m.diskWriteHistory)
}

func (m *Monitor) calculateSpeed(history []SpeedSample) float64 {
	if len(history) < 2 {
		return 0
	}

	now := time.Now()
	cutoff := now.Add(-m.windowSize)

	var totalBytes int64
	var oldestTime time.Time
	var newestTime time.Time

	for _, sample := range history {
		if sample.Timestamp.After(cutoff) {
			if oldestTime.IsZero() || sample.Timestamp.Before(oldestTime) {
				oldestTime = sample.Timestamp
			}
			if newestTime.IsZero() || sample.Timestamp.After(newestTime) {
				newestTime = sample.Timestamp
			}
			totalBytes += sample.Bytes
		}
	}

	if oldestTime.IsZero() || newestTime.IsZero() {
		return 0
	}

	duration := newestTime.Sub(oldestTime).Seconds()
	if duration <= 0 {
		return 0
	}

	return float64(totalBytes) / duration / 1024 / 1024
}

func (m *Monitor) GetTotalStats() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]float64{
		"send_speed_mbps":     m.calculateSpeed(m.sendHistory),
		"receive_speed_mbps":  m.calculateSpeed(m.receiveHistory),
		"disk_read_mbps":      m.calculateSpeed(m.diskReadHistory),
		"disk_write_mbps":     m.calculateSpeed(m.diskWriteHistory),
	}
}

func RecordSend(bytes int64) {
	GetMonitor().RecordSend(bytes)
}

func RecordReceive(bytes int64) {
	GetMonitor().RecordReceive(bytes)
}

func RecordDiskRead(bytes int64) {
	GetMonitor().RecordDiskRead(bytes)
}

func RecordDiskWrite(bytes int64) {
	GetMonitor().RecordDiskWrite(bytes)
}
