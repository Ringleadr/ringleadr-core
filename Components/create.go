package Components

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"strconv"
	"strings"
)

func StartComponent(comp *Datatypes.Component, appName string) error {
	if comp.Name == "" {
		comp.Name = comp.Image
	}

	if !strings.Contains(comp.Image, "/") {
		comp.Image = fmt.Sprintf("docker.io/library/%s", comp.Image)
	}

	runtime := Containers.GetContainerRuntime()
	for i := 0; i < comp.Replicas; i++ {
		cont := &Containers.Container{
			Name:  Containers.GetContainerNameForComponent(comp.Name, appName, i),
			Image: comp.Image,
			Labels: map[string]string{
				"agogos.managed":  "",
				"agogos.owned.by": appName,
				fmt.Sprintf("agogos.%s.%s.replica", appName, comp.Name): strconv.Itoa(i),
			},
		}
		err := runtime.CreateContainer(cont)
		if err != nil {
			return err
		}
	}
	return nil
}
