package seed

import (
	"codegym/backend/internal/model"
)

// GoBasics returns the 50-question Go basic bank (基础题), loaded from
// data/bank/go.json.
func GoBasics() ([]model.Exercise, error) {
	return buildBank("", "go")
}
