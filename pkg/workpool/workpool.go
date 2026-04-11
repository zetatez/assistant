package workpool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type WorkPool struct {
	name          string
	maxConcurrent int
	maxQueue      int
	sem           chan struct{}
	tasks         chan func()
	wg            sync.WaitGroup
	taskWg        sync.WaitGroup
	running       int32
	rejected      int64
	queued        int64
	currentQueued int64
}

func New(name string, maxConcurrent, maxQueue int) *WorkPool {
	return &WorkPool{
		name:          name,
		maxConcurrent: maxConcurrent,
		maxQueue:      maxQueue,
		sem:           make(chan struct{}, maxConcurrent),
		tasks:         make(chan func(), maxQueue),
	}
}

func (p *WorkPool) Go(task func()) bool {
	select {
	case p.tasks <- task:
		atomic.AddInt64(&p.queued, 1)
		atomic.AddInt64(&p.currentQueued, 1)
		return true
	default:
		atomic.AddInt64(&p.rejected, 1)
		return false
	}
}

func (p *WorkPool) GoCtx(ctx context.Context, task func()) bool {
	select {
	case <-ctx.Done():
		return false
	case p.tasks <- task:
		atomic.AddInt64(&p.queued, 1)
		atomic.AddInt64(&p.currentQueued, 1)
		return true
	default:
		atomic.AddInt64(&p.rejected, 1)
		return false
	}
}

func (p *WorkPool) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *WorkPool) worker() {
	defer p.wg.Done()
	for task := range p.tasks {
		atomic.AddInt64(&p.currentQueued, -1)
		p.sem <- struct{}{}
		atomic.AddInt32(&p.running, 1)
		p.taskWg.Add(1)
		go func() {
			defer p.taskWg.Done()
			task()
			atomic.AddInt32(&p.running, -1)
			<-p.sem
		}()
	}
}

func (p *WorkPool) Stop() {
	close(p.tasks)
	p.wg.Wait()
	p.taskWg.Wait()
}

func (p *WorkPool) Wait() {
	p.wg.Wait()
}

func (p *WorkPool) Stats() (running, currentQueued, totalQueued, rejected int64) {
	return int64(atomic.LoadInt32(&p.running)), atomic.LoadInt64(&p.currentQueued), atomic.LoadInt64(&p.queued), atomic.LoadInt64(&p.rejected)
}

func (p *WorkPool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *WorkPool) String() string {
	return p.name
}

type Config struct {
	Name          string
	MaxConcurrent int
	MaxQueue      int
	WorkerCount   int
}

func NewWithConfig(cfg Config) *WorkPool {
	p := New(cfg.Name, cfg.MaxConcurrent, cfg.MaxQueue)
	if cfg.WorkerCount > 0 {
		p.Start(cfg.WorkerCount)
	}
	return p
}

type TimedWorkPool struct {
	pool    *WorkPool
	timeout time.Duration
}

func NewTimed(name string, maxConcurrent, maxQueue int, timeout time.Duration) *TimedWorkPool {
	return &TimedWorkPool{
		pool:    New(name, maxConcurrent, maxQueue),
		timeout: timeout,
	}
}

func (p *TimedWorkPool) Go(task func()) bool {
	return p.pool.GoCtx(context.Background(), func() {
		ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
		defer cancel()
		taskWithContext(ctx, task)
	})
}

func taskWithContext(ctx context.Context, task func()) {
	select {
	case <-ctx.Done():
	default:
		task()
	}
}

func (p *TimedWorkPool) Start(workerCount int) {
	p.pool.Start(workerCount)
}

func (p *TimedWorkPool) Stop() {
	p.pool.Stop()
}

func (p *TimedWorkPool) Stats() (running, currentQueued, totalQueued, rejected int64) {
	return p.pool.Stats()
}
