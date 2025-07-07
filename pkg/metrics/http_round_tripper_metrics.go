package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/guidomantilla/yarumo/pkg/comm"
)

type HttpMetricsRoundTripper struct {
	requestCounter  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	next            http.RoundTripper
}

func NewMetricsRoundTripper(namespace string, next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}

	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "host", "path", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "Histogram of HTTP request durations",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "host", "path", "status"},
	)

	// Register metrics
	prometheus.MustRegister(requestCounter, requestDuration)

	return &HttpMetricsRoundTripper{
		next:            next,
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
	}
}

func (tripper *HttpMetricsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := tripper.next.RoundTrip(req)
	duration := time.Since(start).Seconds()

	status := "error"
	if resp != nil {
		status = http.StatusText(resp.StatusCode)
	}

	labels := prometheus.Labels{
		"method": req.Method,
		"host":   req.URL.Host,
		"path":   comm.NormalizePath(req),
		"status": status,
	}

	tripper.requestCounter.With(labels).Inc()
	tripper.requestDuration.With(labels).Observe(duration)

	return resp, err
}
