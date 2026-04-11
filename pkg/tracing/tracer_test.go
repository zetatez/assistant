package tracing

import (
	"context"
	"testing"
)

func TestTracer_GenerateIDs(t *testing.T) {
	tid := GenerateTraceID()
	if len(tid) != 32 {
		t.Errorf("expected trace ID length 32, got %d", len(tid))
	}

	sid := GenerateSpanID()
	if len(sid) != 16 {
		t.Errorf("expected span ID length 16, got %d", len(sid))
	}
}

func TestTracer_ShouldSample_Disabled(t *testing.T) {
	tracer := &Tracer{enabled: false, sampleRate: 1.0}
	if tracer.ShouldSample() {
		t.Error("expected false when disabled")
	}
}

func TestTracer_ShouldSample_ZeroRate(t *testing.T) {
	tracer := &Tracer{enabled: true, sampleRate: 0.0}
	if tracer.ShouldSample() {
		t.Error("expected false when sample rate is 0")
	}
}

func TestTracer_ShouldSample_FullRate(t *testing.T) {
	tracer := &Tracer{enabled: true, sampleRate: 1.0}
	if !tracer.ShouldSample() {
		t.Error("expected true when sample rate is 1.0")
	}
}

func TestTracer_StartEndSpan(t *testing.T) {
	tracer := &Tracer{enabled: true, sampleRate: 1.0, spans: make([]Span, 0)}
	ctx := context.Background()

	_, span := tracer.StartSpan(ctx, "test-operation")
	if span == nil {
		t.Skip("span is nil when not sampled")
	}
	if span.Operation != "test-operation" {
		t.Errorf("expected operation 'test-operation', got '%s'", span.Operation)
	}

	tracer.EndSpan(span)
	if span.EndTime.IsZero() {
		t.Error("expected EndTime to be set after EndSpan")
	}
}

func TestTracer_ParentChild(t *testing.T) {
	tracer := &Tracer{enabled: true, sampleRate: 1.0, spans: make([]Span, 0)}
	ctx := context.Background()

	ctx, parent := tracer.StartSpan(ctx, "parent")
	tracer.EndSpan(parent)

	if parent == nil {
		t.Skip("parent span is nil when not sampled")
	}

	parentTraceID := parent.TraceID

	ctx2 := InjectContext(ctx, parent)
	_, child := tracer.StartSpan(ctx2, "child")

	if child == nil {
		t.Skip("child span is nil when not sampled")
	}

	if child.TraceID != parentTraceID {
		t.Errorf("child trace ID should match parent: got %s, want %s", child.TraceID, parentTraceID)
	}
	if child.ParentID != parent.SpanID {
		t.Errorf("child parent ID should be parent span ID: got %s, want %s", child.ParentID, parent.SpanID)
	}
}

func TestTracer_ContextPropagation(t *testing.T) {
	tracer := &Tracer{enabled: true, sampleRate: 1.0, spans: make([]Span, 0)}
	ctx := context.Background()

	ctx, span := tracer.StartSpan(ctx, "test")
	if span == nil {
		t.Skip("span is nil when not sampled")
	}
	injected := InjectContext(ctx, span)

	extracted := FromContext(injected)
	if extracted == nil {
		t.Error("expected to extract span from context")
	}
	if extracted.SpanID != span.SpanID {
		t.Errorf("expected same span ID")
	}
}

func TestTracer_AddTag(t *testing.T) {
	tracer := &Tracer{enabled: true, sampleRate: 1.0, spans: make([]Span, 0)}
	ctx := context.Background()
	_, span := tracer.StartSpan(ctx, "test")

	tracer.AddTag(span, "key", "value")
	if span.Tags["key"] != "value" {
		t.Errorf("expected tag 'value', got '%s'", span.Tags["key"])
	}
}

func TestTracer_Log(t *testing.T) {
	tracer := &Tracer{enabled: true, sampleRate: 1.0, spans: make([]Span, 0)}
	ctx := context.Background()
	_, span := tracer.StartSpan(ctx, "test")

	tracer.Log(span, "event", "something happened")
	if len(span.Logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(span.Logs))
	}
}

func TestTracer_GetSpans(t *testing.T) {
	tracer := &Tracer{enabled: true, sampleRate: 1.0, spans: make([]Span, 0)}
	ctx := context.Background()

	_, span1 := tracer.StartSpan(ctx, "op1")
	tracer.EndSpan(span1)
	_, span2 := tracer.StartSpan(ctx, "op2")
	tracer.EndSpan(span2)

	spans := tracer.GetSpans()
	if len(spans) != 2 {
		t.Errorf("expected 2 spans, got %d", len(spans))
	}
}
