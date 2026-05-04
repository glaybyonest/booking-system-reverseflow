package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests.",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	BookingHoldTotal         = promauto.NewCounter(prometheus.CounterOpts{Name: "booking_hold_total", Help: "Successful booking holds."})
	BookingHoldFailedTotal   = promauto.NewCounter(prometheus.CounterOpts{Name: "booking_hold_failed_total", Help: "Failed booking holds."})
	BookingConfirmedTotal    = promauto.NewCounter(prometheus.CounterOpts{Name: "booking_confirmed_total", Help: "Confirmed bookings."})
	BookingExpiredTotal      = promauto.NewCounter(prometheus.CounterOpts{Name: "booking_expired_total", Help: "Expired bookings."})
	PaymentSucceededTotal    = promauto.NewCounter(prometheus.CounterOpts{Name: "payment_succeeded_total", Help: "Successful payments."})
	PaymentFailedTotal       = promauto.NewCounter(prometheus.CounterOpts{Name: "payment_failed_total", Help: "Failed payments."})
	OutboxPublishedTotal     = promauto.NewCounter(prometheus.CounterOpts{Name: "outbox_published_total", Help: "Published outbox events."})
	NotificationCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{Name: "notification_created_total", Help: "Created notifications."})
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		path := r.URL.Path
		HTTPRequestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(rec.status)).Inc()
		HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
	})
}

func Handler() http.Handler {
	return promhttp.Handler()
}
