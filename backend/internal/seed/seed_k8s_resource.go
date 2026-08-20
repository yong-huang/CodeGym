package seed

import (
	"codegym/backend/internal/model"
)

// K8sResourceConfig returns the Kubernetes "资源配置" series (30 questions),
// loaded from data/bank/kubernetes-resource.json.
func K8sResourceConfig() ([]model.Exercise, error) {
	return buildBank("", "kubernetes-resource")
}
