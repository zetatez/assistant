package cron

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type CronMgr struct {
	c            *cron.Cron
	tasks        map[string]*Task
	mu           sync.RWMutex
	maxHistCount int
	logger       *log.Logger
}

func New(opts ...Option) *CronMgr {
	m := &CronMgr{
		c: cron.New(
			cron.WithSeconds(),
			cron.WithChain(cron.Recover(cron.DefaultLogger)),
		),
		tasks:        make(map[string]*Task),
		maxHistCount: 10,
		logger:       log.Default(),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

type Option func(*CronMgr)

func WithLogger(logger *log.Logger) Option {
	return func(m *CronMgr) {
		if logger != nil {
			m.logger = logger
		}
	}
}

func WithMaxHistory(count int) Option {
	return func(m *CronMgr) {
		if count > 0 {
			m.maxHistCount = count
		}
	}
}

type Task struct {
	Name       string
	Schedule   string
	Job        TaskFunc
	EntryID    cron.EntryID
	ResultHist []TaskResult
}

func (t *Task) LastResult() *TaskResult {
	if len(t.ResultHist) == 0 {
		return nil
	}
	return &t.ResultHist[len(t.ResultHist)-1]
}

func (t *Task) IsRunning() bool {
	return t.EntryID != 0
}

type TaskFunc func() error

type TaskResult struct {
	StartTime time.Time
	Duration  time.Duration
	Success   bool
	ErrorMsg  string
}

func (m *CronMgr) AddTask(name string, spec string, fn TaskFunc) error {
	if name == "" {
		return fmt.Errorf("task name cannot be empty")
	}
	if fn == nil {
		return fmt.Errorf("task function cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[name]; exists {
		return fmt.Errorf("task %q already exists", name)
	}

	if err := m.validateSpec(spec); err != nil {
		return fmt.Errorf("invalid cron spec: %w", err)
	}

	task := &Task{
		Name:     name,
		Schedule: spec,
		Job:      fn,
	}

	wrapped := m.wrapTask(name, fn)

	id, err := m.c.AddFunc(spec, wrapped)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	task.EntryID = id
	m.tasks[name] = task
	m.logger.Printf("[cron] task %q added with schedule %q", name, spec)
	return nil
}

func (m *CronMgr) validateSpec(spec string) error {
	_, err := cron.ParseStandard(spec)
	if err != nil {
		return err
	}
	return nil
}

func (m *CronMgr) wrapTask(name string, fn TaskFunc) func() {
	return func() {
		start := time.Now()
		success := true
		var errMsg string

		defer func() {
			if r := recover(); r != nil {
				success = false
				errMsg = fmt.Sprintf("panic: %v", r)
			}

			duration := time.Since(start)
			result := TaskResult{
				StartTime: start,
				Duration:  duration,
				Success:   success,
				ErrorMsg:  errMsg,
			}

			m.mu.Lock()
			defer m.mu.Unlock()

			task, ok := m.tasks[name]
			if !ok {
				return
			}

			task.ResultHist = append(task.ResultHist, result)
			if len(task.ResultHist) > m.maxHistCount {
				task.ResultHist = task.ResultHist[len(task.ResultHist)-m.maxHistCount:]
			}

			status := "OK"
			if !success {
				status = "FAIL"
			}
			m.logger.Printf("[cron][%s] finished (%s) in %s", name, status, duration)
			if errMsg != "" {
				m.logger.Printf("[cron][%s] error: %s", name, errMsg)
			}
		}()

		if err := fn(); err != nil {
			success = false
			errMsg = err.Error()
		}
	}
}

func (m *CronMgr) RemoveTask(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[name]
	if !ok {
		return fmt.Errorf("task %q not found", name)
	}

	m.c.Remove(task.EntryID)
	delete(m.tasks, name)
	m.logger.Printf("[cron] task %q removed", name)
	return nil
}

func (m *CronMgr) GetTask(name string) (*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[name]
	if !ok {
		return nil, fmt.Errorf("task %q not found", name)
	}
	return task, nil
}

func (m *CronMgr) ListTasks() []Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		list = append(list, *t)
	}
	return list
}

func (m *CronMgr) ListTaskNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.tasks))
	for name := range m.tasks {
		names = append(names, name)
	}
	return names
}

func (m *CronMgr) ListResults(name string) ([]TaskResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[name]
	if !ok {
		return nil, fmt.Errorf("task %q not found", name)
	}

	hist := make([]TaskResult, len(task.ResultHist))
	copy(hist, task.ResultHist)
	return hist, nil
}

func (m *CronMgr) RunTaskNow(name string) error {
	m.mu.RLock()
	task, ok := m.tasks[name]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("task %q not found", name)
	}

	if task.EntryID != 0 {
		m.c.Entry(task.EntryID).Job.Run()
	}
	return nil
}

func (m *CronMgr) NextRunTime(name string) (time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[name]
	if !ok {
		return time.Time{}, fmt.Errorf("task %q not found", name)
	}

	entry := m.c.Entry(task.EntryID)
	if entry.Next.IsZero() {
		return time.Time{}, fmt.Errorf("task %q has no scheduled next run", name)
	}
	return entry.Next, nil
}

func (m *CronMgr) TaskCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tasks)
}

func (m *CronMgr) Start() {
	m.c.Start()
	m.logger.Println("[cron] started")
}

func (m *CronMgr) StartWithContext(ctx context.Context) {
	m.c.Start()
	go func() {
		<-ctx.Done()
		m.Stop()
	}()
	m.logger.Println("[cron] started with context")
}

func (m *CronMgr) Stop() {
	ctx := m.c.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
		m.logger.Println("[cron] stop timeout")
	}
	m.logger.Println("[cron] stopped")
}

func (m *CronMgr) StopWithTimeout(timeout time.Duration) {
	ctx := m.c.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(timeout):
		m.logger.Println("[cron] stop timeout")
	}
	m.logger.Println("[cron] stopped")
}
