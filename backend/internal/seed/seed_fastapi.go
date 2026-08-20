package seed

import (
	"codegym/backend/internal/model"
)

// FastAPIBasics returns the 30-question FastAPI tutorial bank, which is part
// of the Python stack (subcategory "FastAPI"; answer-match judging). Loaded
// from data/bank/fastapi.json.
func FastAPIBasics() ([]model.Exercise, error) {
	return buildBank("", "fastapi")
}
