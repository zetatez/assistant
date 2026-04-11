package metrics

import (
	"testing"
)

func TestCounter_Inc(t *testing.T) {
	c := &Counter{}
	c.Inc()
	if c.Value() != 1 {
		t.Errorf("expected 1, got %d", c.Value())
	}
	c.Inc()
	c.Inc()
	if c.Value() != 3 {
		t.Errorf("expected 3, got %d", c.Value())
	}
}

func TestCounter_Add(t *testing.T) {
	c := &Counter{}
	c.Add(5)
	if c.Value() != 5 {
		t.Errorf("expected 5, got %d", c.Value())
	}
}

func TestGauge_Set(t *testing.T) {
	g := &Gauge{}
	g.Set(100)
	if g.Value() != 100 {
		t.Errorf("expected 100, got %d", g.Value())
	}
}

func TestGauge_IncDec(t *testing.T) {
	g := &Gauge{}
	g.Inc()
	g.Inc()
	g.Dec()
	if g.Value() != 1 {
		t.Errorf("expected 1, got %d", g.Value())
	}
}

func TestGauge_Add(t *testing.T) {
	g := &Gauge{}
	g.Add(50)
	g.Add(-30)
	if g.Value() != 20 {
		t.Errorf("expected 20, got %d", g.Value())
	}
}

func TestHistogram_Observe(t *testing.T) {
	h := NewHistogram()

	h.Observe(0.05)
	h.Observe(0.3)
	h.Observe(0.8)
	h.Observe(2.0)

	if h.Count() != 4 {
		t.Errorf("expected count 4, got %d", h.Count())
	}

	sum := h.Sum()
	if sum < 3.14 || sum > 3.16 {
		t.Errorf("expected sum ~3.15, got %f", sum)
	}

	avg := h.Avg()
	if avg < 0.78 || avg > 0.80 {
		t.Errorf("expected avg ~0.79, got %f", avg)
	}
}

func TestHistogram_Empty(t *testing.T) {
	h := NewHistogram()

	if h.Count() != 0 {
		t.Errorf("expected count 0, got %d", h.Count())
	}
	if h.Avg() != 0 {
		t.Errorf("expected avg 0, got %f", h.Avg())
	}
}

func TestTimer_ObserveDuration(t *testing.T) {
	h := NewHistogram()
	timer := NewTimer(h)

	timer.ObserveDuration()

	if h.Count() != 1 {
		t.Errorf("expected 1 observation, got %d", h.Count())
	}
}

func TestFormatPrometheus(t *testing.T) {
	RequestsTotal = &Counter{}
	ErrorsTotal = &Counter{}
	ActiveRequests = &Gauge{}
	RequestsDuration = NewHistogram()
	DBQueryDuration = NewHistogram()
	LLMCallDuration = NewHistogram()

	output := FormatPrometheus()
	if output == "" {
		t.Error("expected non-empty output")
	}
}
