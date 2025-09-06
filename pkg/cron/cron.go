package cron

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type TaskFunc func() error

type TaskResult struct {
	StartTime time.Time
	Duration  time.Duration
	Success   bool
	ErrorMsg  string
}

type Task struct {
	Name       string
	Schedule   string
	Job        TaskFunc
	EntryID    cron.EntryID
	ResultHist []TaskResult
}

type CronMgr struct {
	c            *cron.Cron
	tasks        map[string]*Task
	mu           sync.RWMutex
	maxHistCount int // 任务保留的最大执行历史数
	logger       *log.Logger
}

func NewCronMgr(logger *log.Logger) *CronMgr {
	if logger == nil {
		logger = log.Default()
	}
	return &CronMgr{
		c:            cron.New(cron.WithSeconds()),
		tasks:        make(map[string]*Task),
		logger:       logger,
		maxHistCount: 10,
	}
}

func (m *CronMgr) AddTask(name string, spec string, fn TaskFunc) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[name]; exists {
		return fmt.Errorf("task %q already exists", name)
	}

	task := &Task{
		Name:     name,
		Schedule: spec,
		Job:      fn,
	}

	wrapped := func() {
		start := time.Now()
		success := true
		var errMsg string

		defer func() {
			if r := recover(); r != nil {
				success = false
				errMsg = fmt.Sprintf("panic: %v", r)
			}
			duration := time.Since(start)

			m.mu.Lock()
			defer m.mu.Unlock()

			task.ResultHist = append(task.ResultHist, TaskResult{
				StartTime: start,
				Duration:  duration,
				Success:   success,
				ErrorMsg:  errMsg,
			})
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

	id, err := m.c.AddFunc(spec, wrapped)
	if err != nil {
		return err
	}

	task.EntryID = id
	m.tasks[name] = task
	m.logger.Printf("[cron] task %q added with schedule %q", name, spec)
	return nil
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

func (m *CronMgr) ListTasks() []Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []Task
	for _, t := range m.tasks {
		list = append(list, *t)
	}
	return list
}

func (m *CronMgr) ListResults(name string) ([]TaskResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	task, ok := m.tasks[name]
	if !ok {
		return nil, fmt.Errorf("task %q not found", name)
	}
	return append([]TaskResult(nil), task.ResultHist...), nil
}

func (m *CronMgr) Start() {
	m.c.Start()
	m.logger.Println("[cron] started")
}

func (m *CronMgr) Stop() {
	ctx := m.c.Stop()
	<-ctx.Done()
	m.logger.Println("[cron] stopped")
}
