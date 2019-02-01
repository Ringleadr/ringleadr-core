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

	var networkNames = []string{appName}
	for _, net := range networks {
		networkNames = append(networkNames, fmt.Sprintf("agogos-%s", net))
	}

	var formatNetworks []Containers.Network
	for _, net := range networkNames {
		formatNetworks = append(formatNetworks, Containers.Network{Name: net})
	}

	runtime := Containers.GetContainerRuntime()
	var storage []Containers.StorageMount
	for _, s := range comp.Storage {
		storage = append(storage, Containers.StorageMount{Name: s.Name, MountPath: s.MountPath})
	}

	var env []string
	if Containers.UseProxy {
		env = append(comp.Env, "HTTP_PROXY=http://agogos-proxy:8888",
			"HTTPS_PROXY=http://agogos-proxy:8888", "http_proxy=http://agogos-proxy:8888",
			"https_proxy=http://agogos-proxy:8888")
	} else {
		env = comp.Env
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
			Networks: formatNetworks,
			Alias:    origName,
			Env:      env,
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

	var networkNames = []string{appName}
	for _, net := range networks {
		networkNames = append(networkNames, fmt.Sprintf("agogos-%s", net))
	}

	var formatNetworks []Containers.Network
	for _, net := range networkNames {
		formatNetworks = append(formatNetworks, Containers.Network{Name: net})
	}

	runtime := Containers.GetContainerRuntime()
	var storage []Containers.StorageMount
	for _, s := range comp.Storage {
		storage = append(storage, Containers.StorageMount{Name: s.Name, MountPath: s.MountPath})
	}

	var env []string
	if Containers.UseProxy {
		env = append(comp.Env, "HTTP_PROXY=http://agogos-proxy:8888",
			"HTTPS_PROXY=http://agogos-proxy:8888", "http_proxy=http://agogos-proxy:8888",
			"https_proxy=http://agogos-proxy:8888")
	} else {
		env = comp.Env
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
		Networks: formatNetworks,
		Alias:    origName,
		Env:      env,
	}
	err := runtime.CreateContainer(cont)
	if err != nil {
		return err
	}
	return nil
}
