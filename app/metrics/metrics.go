package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

//revive:disable:exported

var (
	probabilityLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "probability_latency",
			Help:    "Probability calculate latency (histogram)",
			Buckets: []float64{.0001, .0005, .005, .015, .1},
		},
		[]string{"step"},
	)
)

func ProbabilityLatency(step string, duration float64) {
	probabilityLatency.WithLabelValues(step).Observe(duration)
}

func init() {
	prometheus.MustRegister(probabilityLatency)
}
