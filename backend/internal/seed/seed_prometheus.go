package seed

import (
	"codegym/backend/internal/model"
)

// PrometheusBasics returns the 30-question Prometheus bank (promtool/API +
// PromQL + YAML), loaded from data/bank/prometheus.json.
func PrometheusBasics() ([]model.Exercise, error) {
	return buildBank("", "prometheus")
}
