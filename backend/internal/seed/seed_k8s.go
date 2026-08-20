package seed

import (
	"codegym/backend/internal/model"
)

// K8sBasics returns the 50-question Kubernetes basic bank (基础题), loaded
// from data/bank/kubernetes.json.
func K8sBasics() ([]model.Exercise, error) {
	return buildBank("", "kubernetes")
}
