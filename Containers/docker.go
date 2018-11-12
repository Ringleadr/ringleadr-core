//Docker implementation for the ContainerRuntime interface
package Containers

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"io"
	"os"
)

type DockerRuntime struct{}

func (DockerRuntime) AssertOnline() error {
	cli := GetDockerClient()
	ctx := context.Background()

	_, err := cli.ServerVersion(ctx)
	if err != nil {
		panic("Agogos requires Docker to be running to start")
	}

	return nil
}

func (DockerRuntime) CreateContainer(cont *Container) error {
	cli := GetDockerClient()
	ctx := context.Background()

	reader, err := cli.ImagePull(ctx, cont.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, reader)

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
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        cont.Image,
		Labels:       cont.Labels,
		Env:          cont.Env,
		ExposedPorts: ports,
	}, &container.HostConfig{
		PortBindings: portBind,
	}, nil, cont.Name)
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
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
	//TODO try to get env here
	cont := &Container{
		ID:     dockerCont.ID,
		Image:  dockerCont.Image,
		Name:   dockerCont.Names[0],
		Labels: dockerCont.Labels,
		Status: fmt.Sprintf("%s: %s", dockerCont.State, dockerCont.Status),
	}
	return cont, nil
}

//TODO handle ports when in interface
