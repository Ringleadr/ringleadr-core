package Components

import (
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Datatypes"
)

func StartComponent(comp *Datatypes.Component) error {
	cont := &Containers.Container{
		Name: comp.Name,
		Image: comp.Image,
		Labels: map[string]string{
			"agogos.managed": "",
		},
	}
	runtime := Containers.GetContainerRuntime()
	return runtime.CreateContainer(cont)
}
