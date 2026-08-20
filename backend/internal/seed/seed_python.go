package seed

import (
	"codegym/backend/internal/model"
)

// PythonBasics returns the 50-question Python basic bank (基础), loaded from
// data/bank/python.json.
func PythonBasics() ([]model.Exercise, error) {
	return buildBank("", "python")
}
