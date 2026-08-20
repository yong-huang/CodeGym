package seed

import (
	"codegym/backend/internal/model"
)

// GoConcurrency returns the 50-question Go concurrency bank (并发), loaded
// from data/bank/go-concurrency.json.
func GoConcurrency() ([]model.Exercise, error) {
	return buildBank("", "go-concurrency")
}
