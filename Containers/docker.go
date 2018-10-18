//Docker implementation for the ContainerRuntime interface
package Containers

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/pkg/errors"
	"io"
	"os"
)

type DockerRuntime struct{}

func (DockerRuntime) CreateContainer(cont *Container) error {
	cli := GetDockerClient()
	ctx := context.Background()

	reader, err := cli.ImagePull(ctx, cont.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: cont.Image,
		Labels: cont.Labels,
	}, nil, nil, cont.Name)
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	return nil
}

func (DockerRuntime) ReadContainer(id string) (*Container, error) {
	//TODO Implement
	panic("implement me")
}

func (DockerRuntime) ReadAllContainers() ([]*Container, error) {
	//TODO Implement
	cli := GetDockerClient()

	filter := filters.NewArgs()

	//Only list the containers we are managing
	filter.Add("label", "agogos.managed")
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{Filters:filter})
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
	for _, container := range containers {
		newCont, err := dockerContainerToInterface(container)
		if err != nil {
			return nil, err
		}
		returnContainers = append(returnContainers, newCont)
	}
	return returnContainers, nil
}

func dockerContainerToInterface(container types.Container) (*Container, error) {
	if len(container.Names) < 0 {
		return nil, errors.New(fmt.Sprintf("Container %s has no name", container.ID))
	}
	cont := &Container{
		ID:     container.ID,
		Image:  container.Image,
		Name:   container.Names[0],
		Labels: container.Labels,
	}
	return cont, nil
}

//TODO handle ports when in interface
