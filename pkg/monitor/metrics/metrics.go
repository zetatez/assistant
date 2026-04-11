package metrics

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"
)

type Counter struct {
	value uint64
}

func (c *Counter) Inc() {
	atomic.AddUint64(&c.value, 1)
}

func (c *Counter) Add(v uint64) {
	atomic.AddUint64(&c.value, v)
}

func (c *Counter) Value() uint64 {
	return atomic.LoadUint64(&c.value)
}

type Gauge struct {
	value int64
}

func (g *Gauge) Set(v int64) {
	atomic.StoreInt64(&g.value, v)
}

func (g *Gauge) Inc() {
	atomic.AddInt64(&g.value, 1)
}

func (g *Gauge) Dec() {
	atomic.AddInt64(&g.value, -1)
}

func (g *Gauge) Add(v int64) {
	atomic.AddInt64(&g.value, v)
}

func (g *Gauge) Value() int64 {
	return atomic.LoadInt64(&g.value)
}

type Histogram struct {
	count   atomic.Uint64
	sumBits atomic.Uint64
}

func NewHistogram() *Histogram {
	return &Histogram{}
}

func (h *Histogram) Observe(v float64) {
	h.count.Add(1)
	for {
		oldBits := h.sumBits.Load()
		newBits := math.Float64bits(math.Float64frombits(oldBits) + v)
		if h.sumBits.CompareAndSwap(oldBits, newBits) {
			return
		}
	}
}

func (h *Histogram) Count() uint64 {
	return h.count.Load()
}

func (h *Histogram) Sum() float64 {
	return math.Float64frombits(h.sumBits.Load())
}

func (h *Histogram) Avg() float64 {
	c := h.count.Load()
	if c == 0 {
		return 0
	}
	return h.Sum() / float64(c)
}

var DefaultBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

var (
	RequestsTotal    *Counter
	RequestsDuration *Histogram
	ErrorsTotal      *Counter
	ActiveRequests   *Gauge
	DBQueryDuration  *Histogram
	LLMCallDuration  *Histogram
)

func init() {
	RequestsTotal = &Counter{}
	RequestsDuration = NewHistogram()
	ErrorsTotal = &Counter{}
	ActiveRequests = &Gauge{}
	DBQueryDuration = NewHistogram()
	LLMCallDuration = NewHistogram()
}

type Timer struct {
	start time.Time
	h     *Histogram
}

func NewTimer(h *Histogram) *Timer {
	return &Timer{start: time.Now(), h: h}
}

func (t *Timer) ObserveDuration() {
	if t.h != nil {
		t.h.Observe(time.Since(t.start).Seconds())
	}
}

func FormatPrometheus() string {
	var result []byte
	result = append(result, "# TYPE http_requests_total counter\n"...)
	result = append(result, "http_requests_total "...)
	result = appendUint64(result, RequestsTotal.Value())
	result = append(result, '\n')

	result = append(result, "# TYPE http_errors_total counter\n"...)
	result = append(result, "http_errors_total "...)
	result = appendUint64(result, ErrorsTotal.Value())
	result = append(result, '\n')

	result = append(result, "# TYPE http_active_requests gauge\n"...)
	result = append(result, "http_active_requests "...)
	result = appendInt64(result, ActiveRequests.Value())
	result = append(result, '\n')

	result = append(result, "# TYPE http_request_duration_seconds histogram\n"...)
	result = append(result, "http_request_duration_seconds_count "...)
	result = appendUint64(result, RequestsDuration.Count())
	result = append(result, '\n')
	result = append(result, "http_request_duration_seconds_sum "...)
	result = appendFloat(result, RequestsDuration.Sum())
	result = append(result, '\n')

	result = append(result, "# TYPE db_query_duration_seconds histogram\n"...)
	result = append(result, "db_query_duration_seconds_count "...)
	result = appendUint64(result, DBQueryDuration.Count())
	result = append(result, '\n')
	result = append(result, "db_query_duration_seconds_sum "...)
	result = appendFloat(result, DBQueryDuration.Sum())
	result = append(result, '\n')

	result = append(result, "# TYPE llm_call_duration_seconds histogram\n"...)
	result = append(result, "llm_call_duration_seconds_count "...)
	result = appendUint64(result, LLMCallDuration.Count())
	result = append(result, '\n')
	result = append(result, "llm_call_duration_seconds_sum "...)
	result = appendFloat(result, LLMCallDuration.Sum())
	result = append(result, '\n')

	return string(result)
}

func appendUint64(b []byte, v uint64) []byte {
	var buf [20]byte
	i := len(buf)
	for v >= 10 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	i--
	buf[i] = byte('0' + v)
	return append(b, buf[i:]...)
}

func appendInt64(b []byte, v int64) []byte {
	if v < 0 {
		b = append(b, '-')
		v = -v
	}
	return appendUint64(b, uint64(v))
}

func appendFloat(b []byte, v float64) []byte {
	return append(b, fmt.Sprintf("%f", v)...)
}
