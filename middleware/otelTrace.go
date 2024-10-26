package middleware

import (
	"context"
	// "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	// "go.opentelemetry.io/otel/trace"
)

// OpenTelemetry middleware to trace incoming requests
func Trace() gin.HandlerFunc {
	tracer := otel.Tracer("pickeep-tracer")

	return func(c *gin.Context) {
		// Create a span for the incoming request
		ctx, span := tracer.Start(context.Background(), c.Request.URL.Path)
		defer span.End()

		// Pass the context down the middleware chain
		c.Request = c.Request.WithContext(ctx)

		// Set start time
		startTime := time.Now()

		// Process the request
		c.Next()

		// Calculate request duration
		duration := time.Since(startTime)

		// Optionally, record some additional information about the request
		// span.SetAttributes(
		// 	otel.AttributeKey("http.method").String(c.Request.Method),
		// 	otel.AttributeKey("http.url").String(c.Request.URL.Path),
		// 	otel.AttributeKey("http.status_code").Int(c.Writer.Status()),
		// 	otel.AttributeKey("http.duration_ms").Int64(duration.Milliseconds()),
		// )
	}
}
