package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type TraceID string
type SpanID string

type Span struct {
	TraceID   TraceID
	SpanID    SpanID
	ParentID  SpanID
	Operation string
	StartTime time.Time
	EndTime   time.Time
	Tags      map[string]string
	Logs      []SpanLog
	err       error
}

type SpanLog struct {
	Timestamp time.Time
	Key       string
	Value     string
}

type Tracer struct {
	enabled    bool
	sampleRate float32
	spans      []Span
	mu         sync.RWMutex
}

var (
	defaultTracer     *Tracer
	defaultTracerOnce sync.Once
)

func Init(enabled bool, sampleRate float32) *Tracer {
	defaultTracerOnce.Do(func() {
		defaultTracer = &Tracer{
			enabled:    enabled,
			sampleRate: sampleRate,
			spans:      make([]Span, 0, 1000),
		}
	})
	return defaultTracer
}

func Default() *Tracer {
	if defaultTracer == nil {
		return Init(false, 1.0)
	}
	return defaultTracer
}

func (t *Tracer) ShouldSample() bool {
	if !t.enabled {
		return false
	}
	if t.sampleRate >= 1.0 {
		return true
	}
	if t.sampleRate <= 0 {
		return false
	}
	var b [4]byte
	rand.Read(b[:])
	return float32(b[0])/256.0 < t.sampleRate
}

func GenerateTraceID() TraceID {
	b := make([]byte, 16)
	rand.Read(b)
	return TraceID(hex.EncodeToString(b))
}

func GenerateSpanID() SpanID {
	b := make([]byte, 8)
	rand.Read(b)
	return SpanID(hex.EncodeToString(b))
}

func (t *Tracer) StartSpan(ctx context.Context, operation string) (context.Context, *Span) {
	if !t.ShouldSample() {
		return ctx, nil
	}

	span := &Span{
		TraceID:   GenerateTraceID(),
		SpanID:    GenerateSpanID(),
		Operation: operation,
		StartTime: time.Now(),
		Tags:      make(map[string]string),
	}

	if parentSpan := FromContext(ctx); parentSpan != nil {
		span.TraceID = parentSpan.TraceID
		span.ParentID = parentSpan.SpanID
	}

	return InjectContext(ctx, span), span
}

func (t *Tracer) EndSpan(span *Span) {
	if span == nil {
		return
	}
	span.EndTime = time.Now()
	t.mu.Lock()
	t.spans = append(t.spans, *span)
	if len(t.spans) > 10000 {
		t.spans = t.spans[len(t.spans)-5000:]
	}
	t.mu.Unlock()
}

func (t *Tracer) AddTag(span *Span, key, value string) {
	if span == nil {
		return
	}
	span.Tags[key] = value
}

func (t *Tracer) Log(span *Span, key, value string) {
	if span == nil {
		return
	}
	span.Logs = append(span.Logs, SpanLog{
		Timestamp: time.Now(),
		Key:       key,
		Value:     value,
	})
}

func (t *Tracer) SetError(span *Span, err error) {
	if span == nil {
		return
	}
	span.err = err
	span.Tags["error"] = "true"
}

func (t *Tracer) GetSpans() []Span {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]Span, len(t.spans))
	copy(result, t.spans)
	return result
}

func (t *Tracer) ClearSpans() {
	t.mu.Lock()
	t.spans = t.spans[:0]
	t.mu.Unlock()
}

type contextKey string

const spanKey contextKey = "span"

func InjectContext(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, spanKey, span)
}

func FromContext(ctx context.Context) *Span {
	if span, ok := ctx.Value(spanKey).(*Span); ok {
		return span
	}
	return nil
}

func TraceIDToString(tid TraceID) string {
	return string(tid)
}

func SpanIDToString(sid SpanID) string {
	return string(sid)
}
