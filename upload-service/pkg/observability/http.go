package observability

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	metricsRegistry = prometheus.NewRegistry()
	httpRequests    = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "upload_http_requests_total",
			Help: "HTTP requests by path, method and status.",
		},
		[]string{"service", "method", "path", "status"},
	)
	httpLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "upload_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		},
		[]string{"service", "method", "path"},
	)
)

func init() {
	metricsRegistry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	metricsRegistry.MustRegister(prometheus.NewGoCollector())
	metricsRegistry.MustRegister(httpRequests, httpLatency)
}

// HTTPMetricsMiddleware records Prometheus counters/histograms for each request.
func HTTPMetricsMiddleware(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := normalizedPath(c.FullPath(), c.Request.URL.Path)
		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		httpRequests.WithLabelValues(service, method, path, status).Inc()
		httpLatency.WithLabelValues(service, method, path).Observe(time.Since(start).Seconds())
	}
}

// HTTPTraceMiddleware starts a server span for each HTTP request.
func HTTPTraceMiddleware(service string) gin.HandlerFunc {
	tracer := otel.Tracer(service + "-http")
	return func(c *gin.Context) {
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		path := normalizedPath(c.FullPath(), c.Request.URL.Path)
		spanName := c.Request.Method + " " + path

		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", path),
			attribute.String("http.target", c.Request.URL.Path),
			attribute.String("net.host", c.Request.Host),
			attribute.String("service.name", service),
		)

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		status := c.Writer.Status()
		span.SetAttributes(attribute.Int("http.status_code", status))
		if status >= http.StatusInternalServerError {
			span.SetStatus(codes.Error, http.StatusText(status))
		} else {
			span.SetStatus(codes.Ok, "OK")
		}
		span.End()
	}
}

// MetricsHandler exposes Prometheus metrics.
func MetricsHandler() http.Handler {
	return promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})
}

func normalizedPath(fullPath, rawPath string) string {
	if fullPath != "" {
		return fullPath
	}
	if rawPath == "" {
		return "/"
	}
	if idx := strings.Index(rawPath, "?"); idx >= 0 {
		return rawPath[:idx]
	}
	return rawPath
}
