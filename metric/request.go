package metric

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	favicon = "/favicon.ico"
)

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	reqCnt *prometheus.CounterVec
	reqDur *prometheus.HistogramVec
}

// NewPrometheus generates a new set of metrics with a certain subsystem name
func NewPrometheus() (*Prometheus, error) {
	reqCnt := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespaceAPI,
			Name:      "requests_total",
			Help:      "How many HTTP requests processed, partitioned by status code and HTTP method",
		},
		[]string{"code", "method", "path"},
	)
	if err := registerCollector(reqCnt); err != nil {
		return nil, err
	}
	reqDur := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespaceAPI,
			Name:      "request_duration_seconds",
			Help:      "The HTTP request latencies in seconds",
		},
		[]string{"code", "method", "path"},
	)
	if err := registerCollector(reqDur); err != nil {
		return nil, err
	}
	return &Prometheus{
		reqCnt: reqCnt,
		reqDur: reqDur,
	}, nil
}

// PrometheusMiddleware creates the prometheus collector and
// defines status handler function for the middleware
func PrometheusMiddleware() (gin.HandlerFunc, error) {
	p, err := NewPrometheus()
	if err != nil {
		return nil, err
	}
	return p.Middleware(), nil
}

// Middleware defines status handler function for middleware
func (p *Prometheus) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == favicon {
			c.Next()
			return
		}
		start := time.Now()
		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		fullPath := c.FullPath()

		p.reqDur.WithLabelValues(status, c.Request.Method, fullPath).Observe(elapsed)
		p.reqCnt.WithLabelValues(status, c.Request.Method, fullPath).Inc()
	}
}
