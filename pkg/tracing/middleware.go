package tracing

import (
	"github.com/gin-gonic/gin"
)

const (
	HeaderTraceID  = "X-Trace-ID"
	HeaderSpanID   = "X-Span-ID"
	HeaderParentID = "X-Parent-ID"
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tracer := Default()
		if !tracer.enabled {
			c.Next()
			return
		}

		traceIDStr := c.GetHeader(HeaderTraceID)
		var traceID TraceID
		if traceIDStr != "" {
			traceID = TraceID(traceIDStr)
		} else {
			traceID = GenerateTraceID()
		}

		spanID := GenerateSpanID()

		c.Set("trace_id", string(traceID))
		c.Set("span_id", string(spanID))

		c.Header(HeaderTraceID, string(traceID))
		c.Header(HeaderSpanID, string(spanID))

		c.Next()

		if len(c.Errors) > 0 {
			tracer.Log(nil, "error", c.Errors.String())
		}
	}
}

func GetTraceID(c *gin.Context) string {
	if v, exists := c.Get("trace_id"); exists {
		return v.(string)
	}
	return ""
}
