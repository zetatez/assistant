package memory

import (
	"sync"
	"time"
)

type chatBuffer struct {
	messages   []ShortTermMessage
	head       int
	count      int
	mu         sync.Mutex
	lastAccess time.Time
}

type bufferValue struct {
	buf *chatBuffer
}

type ShortTerm struct {
	capacity int
	buffers  sync.Map
	initMu   sync.Mutex
}

func NewShortTerm(capacity int) *ShortTerm {
	if capacity <= 0 {
		capacity = DefaultShortTermCapacity
	}
	return &ShortTerm{
		capacity: capacity,
	}
}

func (s *ShortTerm) getOrCreateBuffer(sessionID string) *chatBuffer {
	if v, ok := s.buffers.Load(sessionID); ok {
		return v.(*bufferValue).buf
	}
	s.initMu.Lock()
	if v, ok := s.buffers.Load(sessionID); ok {
		s.initMu.Unlock()
		return v.(*bufferValue).buf
	}
	buf := &chatBuffer{
		messages:   make([]ShortTermMessage, s.capacity),
		lastAccess: time.Now(),
	}
	s.buffers.Store(sessionID, &bufferValue{buf: buf})
	s.initMu.Unlock()
	return buf
}

func (s *ShortTerm) Add(sessionID, role, content string) {
	buf := s.getOrCreateBuffer(sessionID)
	buf.mu.Lock()
	defer buf.mu.Unlock()
	buf.lastAccess = time.Now()
	buf.messages[buf.head] = ShortTermMessage{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	}
	buf.head = (buf.head + 1) % s.capacity
	if buf.count < s.capacity {
		buf.count++
	}
}

func (s *ShortTerm) GetAll(sessionID string) []ShortTermMessage {
	v, ok := s.buffers.Load(sessionID)
	if !ok {
		return nil
	}
	buf := v.(*bufferValue).buf
	buf.mu.Lock()
	defer buf.mu.Unlock()
	buf.lastAccess = time.Now()
	if buf.count == 0 {
		return nil
	}
	result := make([]ShortTermMessage, buf.count)
	for i := 0; i < buf.count; i++ {
		idx := (buf.head - buf.count + i + s.capacity) % s.capacity
		result[i] = buf.messages[idx]
	}
	return result
}

func (s *ShortTerm) Len(sessionID string) int {
	v, ok := s.buffers.Load(sessionID)
	if !ok {
		return 0
	}
	buf := v.(*bufferValue).buf
	buf.mu.Lock()
	defer buf.mu.Unlock()
	return buf.count
}

func (s *ShortTerm) Clear(sessionID string) {
	v, ok := s.buffers.Load(sessionID)
	if !ok {
		return
	}
	buf := v.(*bufferValue).buf
	buf.mu.Lock()
	defer buf.mu.Unlock()
	buf.head = 0
	buf.count = 0
}

func (s *ShortTerm) CleanupOldSessions(maxAge time.Duration) int {
	cutoff := time.Now().Add(-maxAge)
	var toDelete []string
	s.buffers.Range(func(key, value any) bool {
		buf := value.(*bufferValue).buf
		buf.mu.Lock()
		if buf.lastAccess.Before(cutoff) {
			toDelete = append(toDelete, key.(string))
		}
		buf.mu.Unlock()
		return true
	})
	for _, id := range toDelete {
		s.buffers.Delete(id)
	}
	return len(toDelete)
}
