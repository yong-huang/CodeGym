package seed

import (
	"codegym/backend/internal/model"
)

// OpenShiftDiff returns the 30-question OpenShift-vs-Kubernetes bank (对比):
// DeploymentConfig/Routes/BuildConfig/ImageStream/MachineConfig/OLM/SCC 等 OCP
// 特有对象。Loaded from data/bank/openshift-diff.json.
func OpenShiftDiff() ([]model.Exercise, error) {
	return buildBank("", "openshift-diff")
}
