package Components

import (
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Containers"
)

func StartComponent(comp *Datatypes.Component, appName string) error {
	cont := &Containers.Container{
		Name:  comp.Name,
		Image: comp.Image,
		Labels: map[string]string{
			"agogos.managed":  "",
			"agogos.owned.by": appName,
		},
	}
	runtime := Containers.GetContainerRuntime()
	return runtime.CreateContainer(cont)
}
