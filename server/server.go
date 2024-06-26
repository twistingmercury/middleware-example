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
	serviceName    = "server"
	serviceVersion = "0.0.1"
	environment    = "local"
)

func main() {
	ctx := context.Background()

	var loglevel zerolog.Level
	envLvl := os.Getenv("LOG_LEVEL")

	if envLvl == "" {
		loglevel = 3
	} else {
		loglevel, _ = zerolog.ParseLevel(envLvl)
	}

	// 1.Initialize the logging package.
	if err := logging.Initialize(loglevel, os.Stdout, serviceName, serviceVersion, environment); err != nil {
		log.Panicf("failed to initialize logging: %v", err)
	}
	// 2. Initialize the metrics package.
	if err := metrics.InitializeWithPort(ctx, "9091", namespace, serviceName); err != nil {
		logging.Fatal(err, "failed to initialize server metrics")
	}
	// 3.  publish the metrics
	metrics.Publish()

	// 4. Initialize the tracing package.
	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		logging.Fatal(err, "failed to create server trace exporter")
	}
	if err := tracing.Initialize(traceExporter, serviceName, serviceVersion, environment); err != nil {
		logging.Fatal(err, "failed to initialize server tracing")
	}
	// 5. Initialize the middleare package.
	if err := middleware.Initialize(metrics.Registry(), namespace, serviceName); err != nil {
		logging.Fatal(err, "failed to initialize server middleware")
	}

	// Create a new Gin router
	router := gin.New()
	// 6. Create a gin router and invoke `gin.Use(middleware.Telemetry())`.
	router.Use(gin.Recovery(), middleware.Telemetry())
	// Define a simple route
	router.GET("/epoch", func(c *gin.Context) {
		t := epochTime(c.Request.Context())
		c.String(http.StatusOK, "%d", t)
	})

	httpSvr := &http.Server{Addr: ":8080", Handler: router}
	if err := httpSvr.ListenAndServe(); err != nil {
		logging.Fatal(err, "failed to start http server")
	}
}

func epochTime(context context.Context) int64 {
	cCtx, span := tracing.Start(context, "epochTime", trace.SpanKindInternal)
	time.Sleep(time.Duration(randomInt(cCtx)) * time.Millisecond)
	defer span.End()
	span.SetStatus(codes.Ok, "")
	return time.Now().Unix()
}

func randomInt(pCtx context.Context) int {
	_, span := tracing.Start(pCtx, "randomize", trace.SpanKindInternal)
	defer func() {
		span.SetStatus(codes.Ok, "ok")
		span.End()
	}()
	low := 500
	high := 5000
	return rand.Intn(high-low+1) + low
}
