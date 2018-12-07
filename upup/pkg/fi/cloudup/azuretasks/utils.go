package azuretasks

import (
	"k8s.io/kops/upup/pkg/fi"
)

func GetResourceGroupDependency(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, task := range tasks {
		if _, ok := task.(*ResourceGroup); ok {
			deps = append(deps, task)
		}
	}

	return deps
}
