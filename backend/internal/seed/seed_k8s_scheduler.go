package seed

import (
	"codegym/backend/internal/model"
)

// K8sScheduler returns the 30-question Kubernetes scheduler bank (调度),
// loaded from data/bank/kubernetes-scheduler.json.
func K8sScheduler() ([]model.Exercise, error) {
	return buildBank("", "kubernetes-scheduler")
}
