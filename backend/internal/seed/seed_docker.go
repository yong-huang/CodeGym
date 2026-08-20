package seed

import (
	"codegym/backend/internal/model"
)

// DockerBasics returns the 50-question Docker basic bank (命令 + Dockerfile +
// compose), loaded from data/bank/docker.json.
func DockerBasics() ([]model.Exercise, error) {
	return buildBank("", "docker")
}
