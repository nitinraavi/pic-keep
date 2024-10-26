package main

import (
	"fmt"
	"go-jwt/intializers"
	"go-jwt/routes"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"log"
	"context"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"


	"go.opentelemetry.io/otel"
	// "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)
var (
	tracer       trace.Tracer
	otlpEndpoint string
)
func init() {
	intializers.LoadEnvVariables()
	intializers.ConnectToDB()
	intializers.SyncDatabase()
	
	otlpEndpoint = os.Getenv("OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		log.Fatalln("You MUST set OTLP_ENDPOINT env variable!")
	}
	
}

// List of supported exporters
// https://opentelemetry.io/docs/instrumentation/go/exporters/

// Console Exporter, only for testing
func newConsoleExporter() (oteltrace.SpanExporter, error) {
	return stdouttrace.New()
}
// TracerProvider is an OpenTelemetry TracerProvider.
// It provides Tracers to instrumentation so it can trace operational flow through a system.
func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("pickeep"),
		),
	)

	if err != nil {
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}
func main() {
	// Channel to listen for SIGTERM
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	ctx := context.Background()

	// For testing to print out traces to the console
	exp, err := newConsoleExporter()
	// exp, err := newOTLPExporter(ctx)

	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}
	// Create a new tracer provider with a batch span processor and the given exporter.
	tp := newTraceProvider(exp)

	// Handle shutdown properly so nothing leaks.
	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)

	// Finally, set the tracer that can be used for this package.
	tracer = tp.Tracer("pickeep")

	// Register channel to listen for SIGTERM signals
	signal.Notify(sigs, syscall.SIGTERM)

	// Initialize logger using slog with JSON output
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Create a new Gin router
	r := gin.New()

	// Define a health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Apply slogging middleware to log request attributes
	r.Use(sloggin.New(logger))

	// Register authentication routes
	routes.AuthRoutes(r)

	// Run the server in a goroutine so the main thread isn't blocked
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080" // Default to port 8080 if not specified
		}
		if err := r.Run(":" + port); err != nil {
			fmt.Println("Error starting server:", err)
		}
	}()

	fmt.Println("Starting application")

	// Goroutine to catch the signal and handle graceful shutdown
	go func() {
		sig := <-sigs
		fmt.Println("Caught SIGTERM, shutting down gracefully:", sig)
		// Send a signal to shutdown the app
		done <- true
	}()

	// Block until we receive the shutdown signal
	<-done
	fmt.Println("Exiting application")
}


// func main() {
// 	// Create channels for OS signals and application shutdown
// 	sigs := make(chan os.Signal, 1)
// 	done := make(chan bool, 1)

// 	// Register the channel to listen for SIGTERM signals
// 	signal.Notify(sigs, syscall.SIGTERM)

// 	go func() {
// 		sig := <-sigs
// 		fmt.Println("Caught SIGTERM, shutting down gracefully:", sig)
// 		// Complete any outstanding requests, then signal completion
// 		done <- true
// 	}()

// 	fmt.Println("Starting application")

// 	// Wait for the shutdown signal
// 	<-done
// 	fmt.Println("Exiting application")

// 	// Initialize logger using slog with JSON output
// 	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

// 	// Create a new Gin router
// 	r := gin.New()

// 	// Define a health check endpoint
// 	r.GET("/health", func(c *gin.Context) {
// 		c.JSON(200, gin.H{
// 			"status": "healthy",
// 		})
// 	})

// 	// Apply slogging middleware to log request attributes
// 	r.Use(sloggin.New(logger))

// 	// Register authentication routes
// 	routes.AuthRoutes(r)

// 	// Start the server and listen on the specified port
// 	if err := r.Run(":" + os.Getenv("PORT")); err != nil {
// 		fmt.Println("Error starting server:", err)
// 	}
// }



// func jsonLoggerMiddleware() gin.HandlerFunc {
// 	return gin.LoggerWithFormatter(
// 		func(params gin.LogFormatterParams) string {
// 			log := make(map[string]interface{})

// 			log["status_code"] = params.StatusCode
// 			log["path"] = params.Path
// 			log["method"] = params.Method
// 			log["start_time"] = params.TimeStamp.Format("2006/01/02 - 15:04:05")
// 			log["remote_addr"] = params.ClientIP
// 			log["response_time"] = params.Latency.String()
// 			log["message"] = params.ErrorMessage

// 			s, _ := json.Marshal(log)
// 			return string(s) + "\n"
// 		},
// 	)
// }
