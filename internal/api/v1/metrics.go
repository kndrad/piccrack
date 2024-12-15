package v1

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	requestsTotal *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "requests_total",
				Help: "Counter for total requests made with a method and a url.",
			},
			[]string{"method", "url"},
		),
	}

	reg.MustRegister(m.requestsTotal)

	return m
}

func (m *metrics) IncCounter(r *http.Request) {
	if r == nil {
		panic("request cannot be nil")
	}
	m.requestsTotal.With(prometheus.Labels{
		"method": r.Method,
		"url":    r.URL.String(),
	}).Inc()
}

func (m *metrics) WrapHandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.IncCounter(r)
		next.ServeHTTP(w, r)
	}
}
