package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/twistingmercury/middleware"
	"github.com/twistingmercury/telemetry/logging"
	"github.com/twistingmercury/telemetry/metrics"
	"github.com/twistingmercury/telemetry/tracing"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/trace"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	namespace      = "example"
	serviceName    = "test"
	serviceVersion = "0.0.1"
	environment    = "local"
)

func main() {
	ctx := context.Background()
	// 1.Initialize the logging package.
	if err := logging.Initialize(zerolog.DebugLevel, os.Stdout, serviceName, serviceVersion, environment); err != nil {
		log.Panicf("failed to initialize logging: %v", err)
	}
	// 2. Initialize the metrics package.
	if err := metrics.InitializeWithPort(ctx, "9191", namespace, serviceName); err != nil {
		logging.Fatal(err, "failed to initialize metrics")
	}
	// 3.  publish the metrics
	metrics.Publish()

	// 4. Initialize the tracing package.
	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		logging.Fatal(err, "failed to create trace exporter")
	}
	if err := tracing.Initialize(traceExporter, serviceName, serviceVersion, environment); err != nil {
		logging.Fatal(err, "failed to initialize tracing")
	}
	// 5. Initialize the middleare package.
	if err := middleware.Initialize(metrics.Registry(), namespace, serviceName); err != nil {
		logging.Fatal(err, "failed to initialize middleware")
	}

	cChan := make(chan context.Context)
	defer close(cChan)
	go worker(cChan)

	// Create a new Gin router
	router := gin.New()
	// 6. Create a gin router and invoke `gin.Use(middleware.Telemetry())`.
	router.Use(gin.Recovery(), middleware.Telemetry())
	// Define a simple route
	router.GET("/hello", func(c *gin.Context) {
		cChan <- c.Request.Context()
		c.String(http.StatusOK, "Hello, World!")
	})

	httpSvr := &http.Server{Addr: ":8080", Handler: router}
	if err := httpSvr.ListenAndServe(); err != nil {
		logging.Fatal(err, "failed to start http server")
	}
}

func worker(ctxChan chan context.Context) {
	for {
		select {
		case ctx := <-ctxChan:
			func() {
				sCtx := trace.SpanContextFromContext(ctx)
				logging.InfoWithContext(&sCtx, "starting concurrent worker")
				wCtx, span := tracing.Start(ctx, "concurrent worker", trace.SpanKindInternal)
				defer span.End()
				subWorker(wCtx)
				span.SetStatus(codes.Ok, "ok")
			}()
		}
	}
}

func subWorker(parentCtx context.Context) {
	cCtx, span := tracing.Start(parentCtx, "child worker", trace.SpanKindUnspecified)
	defer span.End()
	sCtx := trace.SpanContextFromContext(cCtx)
	logging.InfoWithContext(&sCtx, "starting sub worker")
	span.SetStatus(codes.Ok, "ok")
	low := 10
	high := 200
	randomNum := rand.Intn(high-low+1) + low
	time.Sleep(time.Duration(randomNum) * time.Millisecond)
}
