package performance

import (
	"runtime"
	"sync"
	"time"

	"lanfiletransfertool/pkg/logger"
)

type Monitor struct {
	stats       *Stats
	pool        *ThreadPool
	mu          sync.RWMutex
	stopChan    chan bool
	running     bool
}

type Stats struct {
	CPUUsage         float64
	MemoryUsage      float64
	NetworkSpeed     float64
	ActiveGoroutines int
	Timestamp        time.Time
}

type ThreadPool struct {
	size    int
	active  int
	tasks   chan func()
	mu      sync.Mutex
	running bool
}

func NewMonitor() *Monitor {
	return &Monitor{
		stats:    &Stats{},
		stopChan: make(chan bool),
	}
}

func (m *Monitor) GetStats() *Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.stats.CPUUsage = calculateCPUUsage()
	m.stats.MemoryUsage = float64(memStats.Alloc) / float64(memStats.Sys) * 100
	m.stats.ActiveGoroutines = runtime.NumGoroutine()
	m.stats.Timestamp = time.Now()

	return m.stats
}

func (m *Monitor) InitPool(size int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pool != nil && m.pool.running {
		return nil
	}

	if size <= 0 {
		size = 5
	}

	pool := &ThreadPool{
		size:    size,
		tasks:   make(chan func(), size*2),
		running: true,
	}

	for i := 0; i < size; i++ {
		go pool.worker(i)
	}

	m.pool = pool
	logger.Info("线程池初始化完成，大小: %d", size)
	return nil
}

func (p *ThreadPool) worker(id int) {
	for task := range p.tasks {
		task()
		p.mu.Lock()
		p.active--
		p.mu.Unlock()
	}
}

func (m *Monitor) StopPool() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pool == nil || !m.pool.running {
		return nil
	}

	close(m.pool.tasks)
	m.pool.running = false
	logger.Info("线程池已停止")
	return nil
}

func (m *Monitor) SubmitTask(task func()) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.pool == nil || !m.pool.running {
		return nil
	}

	m.pool.mu.Lock()
	m.pool.active++
	m.pool.mu.Unlock()

	select {
	case m.pool.tasks <- task:
	default:
		m.pool.mu.Lock()
		m.pool.active--
		m.pool.mu.Unlock()
	}

	return nil
}

func calculateCPUUsage() float64 {
	return 0.0
}

func monitorNetworkSpeed() float64 {
	return 0.0
}
