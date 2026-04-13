package performance

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Task 线程池任务
type Task func()

// Pool 线程池结构体
type Pool struct {
	size       int
	taskQueue  chan Task
	workers    []chan struct{}
	stopChan   chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
	running    bool
	taskCount  int64
}

// NewPool 创建新的线程池
func NewPool(size int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		size:      size,
		taskQueue: make(chan Task, size*2),
		stopChan:  make(chan struct{}),
		ctx:       ctx,
		cancel:    cancel,
		running:   false,
	}
}

// Start 启动线程池
func (p *Pool) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("线程池已在运行")
	}

	p.workers = make([]chan struct{}, p.size)
	for i := 0; i < p.size; i++ {
		p.workers[i] = make(chan struct{})
		p.wg.Add(1)
		go p.worker(i, p.workers[i])
	}

	p.running = true
	return nil
}

// Stop 停止线程池
func (p *Pool) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	// 发送停止信号
	p.cancel()
	close(p.stopChan)

	// 等待所有工作线程完成
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 正常停止
	case <-time.After(10 * time.Second):
		// 超时，强制停止
	}

	p.running = false
	return nil
}

// Submit 提交任务到线程池
func (p *Pool) Submit(task Task) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.running {
		return fmt.Errorf("线程池未运行")
	}

	select {
	case p.taskQueue <- task:
		p.mu.Lock()
		p.taskCount++
		p.mu.Unlock()
		return nil
	default:
		return fmt.Errorf("任务队列已满")
	}
}

// SubmitAsync 异步提交任务（不阻塞）
func (p *Pool) SubmitAsync(task Task) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.running {
		return false
	}

	select {
	case p.taskQueue <- task:
		p.mu.Lock()
		p.taskCount++
		p.mu.Unlock()
		return true
	default:
		return false
	}
}

// worker 工作线程
func (p *Pool) worker(id int, stop chan struct{}) {
	defer p.wg.Done()

	for {
		select {
		case task := <-p.taskQueue:
			if task != nil {
				task()
			}
		case <-stop:
			return
		case <-p.ctx.Done():
			return
		}
	}
}

// IsRunning 检查线程池是否运行中
func (p *Pool) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// GetSize 获取线程池大小
func (p *Pool) GetSize() int {
	return p.size
}

// GetTaskCount 获取已处理任务数
func (p *Pool) GetTaskCount() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.taskCount
}

// GetQueueSize 获取当前队列中的任务数
func (p *Pool) GetQueueSize() int {
	return len(p.taskQueue)
}

// GetQueueCapacity 获取队列容量
func (p *Pool) GetQueueCapacity() int {
	return cap(p.taskQueue)
}

// Resize 调整线程池大小（需要重启）
func (p *Pool) Resize(newSize int) error {
	if newSize <= 0 || newSize > 100 {
		return fmt.Errorf("线程池大小必须在1-100之间")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("线程池运行中，无法调整大小，请先停止")
	}

	p.size = newSize
	p.taskQueue = make(chan Task, newSize*2)
	return nil
}
