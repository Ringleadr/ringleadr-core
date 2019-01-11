//Docker implementation for the ContainerRuntime interface
package Containers

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"strings"
)

type DockerRuntime struct{}

func (DockerRuntime) AssertOnline() error {
	cli := GetDockerClient()
	ctx := context.Background()

	_, err := cli.ServerVersion(ctx)
	if err != nil {
		log.Println("error communicating with docker:", err)
		panic("Agogos requires Docker to be running to start")
	}

	return nil
}

func (DockerRuntime) CreateContainer(cont *Container) error {
	cli := GetDockerClient()
	ctx := context.Background()

	if !strings.Contains(cont.Image, "/") {
		cont.Image = fmt.Sprintf("docker.io/library/%s", cont.Image)
	}

	if !strings.Contains(cont.Image, ":") {
		cont.Image = fmt.Sprintf("%s:latest", cont.Image)
	}

	shouldPull := true
	//Check if image exists locally, if it does then don't pull, otherwise do
	if _, _, err := cli.ImageInspectWithRaw(ctx, cont.Image); err == nil {
		shouldPull = false
	} else {
		errString := err.Error()
		if !strings.Contains(errString, "No such image") {
			log.Println("error checking if image exists: ", cont.Image, errString)
		}
	}

	if shouldPull {
		//Container create does not pull missing images, so we force a pull
		reader, err := cli.ImagePull(ctx, cont.Image, types.ImagePullOptions{})
		if err != nil {
			log.Println("Error pulling image: ", err.Error())
			return err
		}
		_, _ = io.Copy(os.Stdout, reader)
	}

	ports := nat.PortSet{}
	portBind := nat.PortMap{}
	for k, v := range cont.Ports {
		p, err := nat.NewPort("tcp", v)
		if err != nil {
			return err
		}
		ports[p] = struct{}{}
		portBind[p] = []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: k}}
	}

	mounts := []mount.Mount{}
	for _, store := range cont.Storage {
		mounts = append(mounts, mount.Mount{Type: mount.TypeVolume,
			Source: fmt.Sprintf("agogos-%s", store.Name), Target: store.MountPath})
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        cont.Image,
		Labels:       cont.Labels,
		Env:          cont.Env,
		ExposedPorts: ports,
	}, &container.HostConfig{
		PortBindings:  portBind,
		Mounts:        mounts,
		RestartPolicy: container.RestartPolicy{Name: "always"},
	}, nil, cont.Name)
	if err != nil {
		log.Println("Error Creating container: ", err.Error())
		return err
	}

	for _, net := range cont.Networks {
		settings := &network.EndpointSettings{}
		if cont.Alias != "" {
			settings.Aliases = []string{cont.Alias}
		}
		if err := cli.NetworkConnect(ctx, net, resp.ID, settings); err != nil {
			log.Println("error attaching container to network", resp.ID, err)
			return err
		}
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Println("Error starting container: ", err.Error())
		return err
	}

	return nil
}

func (DockerRuntime) ReadContainer(id string) (*Container, error) {
	//TODO deal with id or name
	cli := GetDockerClient()
	ctx := context.Background()

	filter := filters.NewArgs()
	filter.Add("name", id)
	cont, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filter,
		All:     true,
	})

	if err != nil {
		log.Println("Error listing containers: ", err.Error())
		return nil, err
	}

	if len(cont) != 1 {
		return nil, errors.New("did not return single container")
	}
	return dockerContainerToInterface(cont[0])
}

func (DockerRuntime) ReadAllContainers() ([]*Container, error) {
	cli := GetDockerClient()

	filter := filters.NewArgs()

	//Only list the containers we are managing
	filter.Add("label", "agogos.managed")
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: filter,
		All:     true,
	})
	if err != nil {
		log.Println("Error listing containers: ", err.Error())
		return nil, err
	}

	return dockerContainersToInterface(containers...)
}

func (DockerRuntime) ReadAllContainersWithFilter(filter map[string]map[string]bool) ([]*Container, error) {
	cli := GetDockerClient()

	dockerFilter := filters.NewArgs()
	dockerFilter.Add("label", "agogos.managed")
	for k, v := range filter {
		for innerk := range v {
			dockerFilter.Add(k, innerk)
		}
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: dockerFilter,
	})
	if err != nil {
		log.Println("Error listing containers: ", err.Error())
		return nil, err
	}
	return dockerContainersToInterface(containers...)
}

func (DockerRuntime) UpdateContainer(container *Container) error {
	//TODO Implement
	panic("implement me")
}

func (DockerRuntime) DeleteContainer(id string) error {
	//TODO Implement
	panic("implement me")
}

func (d DockerRuntime) DeleteContainerWithFilter(filter map[string]map[string]bool) error {
	cli := GetDockerClient()

	containers, err := d.ReadAllContainersWithFilter(filter)
	if err != nil {
		log.Println("Error retrieving containers with filter: ", filter, err.Error())
		return err
	}

	for _, cont := range containers {
		err = cli.ContainerRemove(context.Background(), cont.ID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
		if err != nil {
			log.Println("Error deleting container", err.Error())
			return err
		}
	}
	return nil
}

func (DockerRuntime) CreateStorage(name string) error {
	cli := GetDockerClient()

	_, err := cli.VolumeCreate(context.Background(), volume.VolumeCreateBody{Name: name})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (DockerRuntime) DeleteStorage(name string) error {
	cli := GetDockerClient()

	err := cli.VolumeRemove(context.Background(), name, false)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (DockerRuntime) CreateNetwork(name string) error {
	cli := GetDockerClient()

	_, err := cli.NetworkCreate(context.Background(), name, types.NetworkCreate{})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (DockerRuntime) DeleteNetwork(name string) error {
	cli := GetDockerClient()

	err := cli.NetworkRemove(context.Background(), name)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (DockerRuntime) NetworkExists(name string) (bool, error) {
	cli := GetDockerClient()

	_, err := cli.NetworkInspect(context.Background(), name, types.NetworkInspectOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "No such network") {
			return false, nil
		}
		fmt.Println("Error checking network: ", err)
		return false, err
	}

	return true, nil
}

func dockerContainersToInterface(containers ...types.Container) ([]*Container, error) {
	var returnContainers []*Container
	for _, cont := range containers {
		newCont, err := dockerContainerToInterface(cont)
		if err != nil {
			return nil, err
		}
		returnContainers = append(returnContainers, newCont)
	}
	return returnContainers, nil
}

func dockerContainerToInterface(dockerCont types.Container) (*Container, error) {
	if len(dockerCont.Names) < 0 {
		return nil, errors.New(fmt.Sprintf("Container %s has no name", dockerCont.ID))
	}
	stringPorts := make(map[string]string)
	for _, port := range dockerCont.Ports {
		stringPorts[string(port.PublicPort)] = string(port.PrivatePort)
	}

	//TODO try to get env here
	cont := &Container{
		ID:     dockerCont.ID,
		Image:  dockerCont.Image,
		Name:   dockerCont.Names[0],
		Labels: dockerCont.Labels,
		Status: fmt.Sprintf("%s: %s", dockerCont.State, dockerCont.Status),
		Ports:  stringPorts,
	}
	return cont, nil
}
