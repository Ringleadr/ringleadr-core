package Components

import (
	"github.com/GodlikePenguin/agogos-host/Containers"
)

func DeleteAllComponents(appName string) error {
	filter := map[string]map[string]bool{
		"label": {
			"agogos.owned.by=" + appName: true,
		},
	}
	runtime := Containers.GetContainerRuntime()
	return runtime.DeleteContainerWithFilter(filter)
}
