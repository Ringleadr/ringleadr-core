package Components

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"strconv"
)

func StartComponent(comp *Datatypes.Component, appName string, appCopy int, networks []string) error {
	origName := comp.Name
	if comp.Name == "" {
		comp.Name = comp.Image
	}

	var newNetworks = []string{appName}
	for _, net := range networks {
		newNetworks = append(newNetworks, fmt.Sprintf("agogos-%s", net))
	}

	runtime := Containers.GetContainerRuntime()
	var storage []Containers.StorageMount
	for _, s := range comp.Storage {
		storage = append(storage, Containers.StorageMount{Name: s.Name, MountPath: s.MountPath})
	}
	for i := 0; i < comp.Replicas; i++ {
		cont := &Containers.Container{
			Name:  Containers.GetContainerNameForComponent(comp.Name, appName, appCopy, i),
			Image: comp.Image,
			Labels: map[string]string{
				"agogos.managed":  "",
				"agogos.owned.by": fmt.Sprintf("%s-%d", appName, appCopy),
				fmt.Sprintf("agogos.%s.%d.%s.replica", appName, appCopy, comp.Name): strconv.Itoa(i),
			},
			Storage:  storage,
			Ports:    comp.Ports,
			Networks: newNetworks,
			Alias:    origName,
			Env:      comp.Env,
		}
		err := runtime.CreateContainer(cont)
		if err != nil {
			return err
		}
	}
	return nil
}

func StartComponentReplica(comp *Datatypes.Component, appName string, appCopy int, networks []string, replica int) error {
	origName := comp.Name
	if comp.Name == "" {
		comp.Name = comp.Image
	}

	var newNetworks = []string{appName}
	for _, net := range networks {
		newNetworks = append(newNetworks, fmt.Sprintf("agogos-%s", net))
	}

	runtime := Containers.GetContainerRuntime()
	var storage []Containers.StorageMount
	for _, s := range comp.Storage {
		storage = append(storage, Containers.StorageMount{Name: s.Name, MountPath: s.MountPath})
	}
	cont := &Containers.Container{
		Name:  Containers.GetContainerNameForComponent(comp.Name, appName, appCopy, replica),
		Image: comp.Image,
		Labels: map[string]string{
			"agogos.managed":  "",
			"agogos.owned.by": fmt.Sprintf("%s-%d", appName, appCopy),
			fmt.Sprintf("agogos.%s.%d.%s.replica", appName, appCopy, comp.Name): strconv.Itoa(replica),
		},
		Storage:  storage,
		Ports:    comp.Ports,
		Networks: newNetworks,
		Alias:    origName,
		Env:      comp.Env,
	}
	err := runtime.CreateContainer(cont)
	if err != nil {
		return err
	}
	return nil
}
