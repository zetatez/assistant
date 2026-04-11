package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ActiveRequests.Inc()
		RequestsTotal.Inc()
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		RequestsDuration.Observe(duration)
		ActiveRequests.Dec()

		if len(c.Errors) > 0 {
			ErrorsTotal.Inc()
		}
	}
}

func RecordDBQuery(duration time.Duration) {
	DBQueryDuration.Observe(duration.Seconds())
}

func RecordLLMCall(duration time.Duration) {
	LLMCallDuration.Observe(duration.Seconds())
}
