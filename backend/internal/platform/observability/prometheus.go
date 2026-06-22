package observability

import (
	"net"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

const metricsPath = "/metrics"

var (
	operationCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "hub",
			Name:      "operation_calls_total",
			Help:      "Total Timed() operation invocations",
		},
		[]string{"operation", "result"},
	)
	operationDurationMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "hub",
			Name:      "operation_duration_milliseconds",
			Help:      "Timed() operation duration in milliseconds",
			Buckets:   []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
		},
		[]string{"operation"},
	)
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "hub",
			Name:      "http_requests_total",
			Help:      "Total HTTP requests",
		},
		[]string{"method", "route", "status"},
	)
	httpRequestDurationMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "hub",
			Name:      "http_request_duration_milliseconds",
			Help:      "HTTP request duration in milliseconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
)

func init() {
	prometheus.MustRegister(
		operationCallsTotal,
		operationDurationMs,
		httpRequestsTotal,
		httpRequestDurationMs,
	)
}

func recordPrometheusOperation(operation string, durationMs int64, failed bool) {
	result := "success"
	if failed {
		result = "error"
	}
	operationCallsTotal.WithLabelValues(operation, result).Inc()
	operationDurationMs.WithLabelValues(operation).Observe(float64(durationMs))
}

// MetricsEndpoint — /metrics с опциональным CIDR allowlist (OBS-01).
func MetricsEndpoint(allowedCIDRs []string) fiber.Handler {
	scrape := MetricsHandler()
	nets := parseCIDRs(allowedCIDRs)
	denyAll := len(allowedCIDRs) > 0 && len(nets) == 0

	return func(c fiber.Ctx) error {
		if len(allowedCIDRs) > 0 {
			if denyAll {
				return fiber.NewError(fiber.StatusForbidden, "metrics endpoint forbidden")
			}
			ip := net.ParseIP(c.IP())
			if ip == nil || !ipAllowed(ip, nets) {
				return fiber.NewError(fiber.StatusForbidden, "metrics endpoint forbidden")
			}
		}
		return scrape(c)
	}
}

// MetricsHandler — Prometheus scrape на GET /metrics (OBS-01).
func MetricsHandler() fiber.Handler {
	handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
	return func(c fiber.Ctx) error {
		handler(c.Context())
		return nil
	}
}

// HTTPMetricsMiddleware — count и latency для Prometheus.
func HTTPMetricsMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Path() == metricsPath {
			return c.Next()
		}

		start := time.Now()
		err := c.Next()

		route := c.Path()
		if r := c.Route(); r != nil && r.Path != "" {
			route = r.Path
		}

		status := c.Response().StatusCode()
		if status == 0 {
			status = fiber.StatusOK
		}
		if err != nil {
			if fe, ok := err.(*fiber.Error); ok && fe.Code > 0 {
				status = fe.Code
			} else if status < fiber.StatusBadRequest {
				status = fiber.StatusInternalServerError
			}
		}

		statusLabel := strconv.Itoa(status)
		method := c.Method()
		ms := float64(time.Since(start).Milliseconds())

		httpRequestsTotal.WithLabelValues(method, route, statusLabel).Inc()
		httpRequestDurationMs.WithLabelValues(method, route).Observe(ms)

		return err
	}
}

// ResetPrometheusMetrics — сброс collectors (только tests).
func ResetPrometheusMetrics() {
	operationCallsTotal.Reset()
	operationDurationMs.Reset()
	httpRequestsTotal.Reset()
	httpRequestDurationMs.Reset()
}
