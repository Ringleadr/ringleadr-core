package Components

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Logger"
)

func DeleteAllComponents(appName string, appCopies int) error {
	runtime := Containers.GetContainerRuntime()
	for i := 0; i < appCopies; i++ {
		filter := map[string]map[string]bool{
			"label": {
				fmt.Sprintf("agogos.owned.by=%s-%d", appName, i): true,
			},
		}
		go func(filter map[string]map[string]bool) {
			err := runtime.DeleteContainerWithFilter(filter)
			if err != nil {
				Logger.ErrPrintf("Error deleting all Components for app %s: %s", appName, err.Error())
			}
		}(filter)
		//Ignore any errors and hope they are fixed later
	}
	return nil
}
