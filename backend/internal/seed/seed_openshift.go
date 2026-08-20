package seed

import (
	"codegym/backend/internal/model"
)

// OpenShiftBasics returns the 50-question OpenShift basic bank (oc 命令 +
// YAML), loaded from data/bank/openshift.json.
func OpenShiftBasics() ([]model.Exercise, error) {
	return buildBank("", "openshift")
}
