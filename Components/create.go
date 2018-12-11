package Components

import (
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Containers"
)

func StartComponent(comp *Datatypes.Component, appName string) error {
	//TODO name should include the name of the app to avoid collisions
	cont := &Containers.Container{
		Name:  Containers.GetContainerNameForComponent(comp.Name, appName, 0),
		Image: comp.Image,
		Labels: map[string]string{
			"agogos.managed":  "",
			"agogos.owned.by": appName,
		},
	}
	runtime := Containers.GetContainerRuntime()
	return runtime.CreateContainer(cont)
}
