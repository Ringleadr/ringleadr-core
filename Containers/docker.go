//Docker implementation for the ContainerRuntime interface
package Containers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"io"
	"os"
	"strconv"
	"strings"
)

type DockerRuntime struct{}

func (DockerRuntime) AssertOnline() error {
	cli := GetDockerClient()
	ctx := context.Background()

	_, err := cli.ServerVersion(ctx)
	if err != nil {
		Logger.ErrPrintln("error communicating with docker:", err)
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
			Logger.ErrPrintln("error checking if image exists: ", cont.Image, errString, "Will try to continue")
			//Attempt to continue creating container
		}
	}

	if shouldPull {
		//Container create does not pull missing images, so we force a pull
		reader, err := cli.ImagePull(ctx, cont.Image, types.ImagePullOptions{})
		if err != nil {
			return errors.New("Error pulling image: " + err.Error())
		}
		_, _ = io.Copy(os.Stdout, reader)
	}

	ports := nat.PortSet{}
	portBind := nat.PortMap{}
	for k, v := range cont.Ports {
		p, err := nat.NewPort("tcp", v)
		if err != nil {
			return errors.New("Error creating port: " + err.Error())
		}
		ports[p] = struct{}{}
		if !strings.Contains(k, ":") {
			portBind[p] = []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: k}}
		} else {
			index := strings.Index(k, ":")
			portBind[p] = []nat.PortBinding{{HostIP: k[:index], HostPort: k[index+1:]}}
		}
	}

	mounts := []mount.Mount{}
	for _, store := range cont.Storage {
		mounts = append(mounts, mount.Mount{Type: mount.TypeVolume,
			Source: fmt.Sprintf("agogos-%s", store.Name), Target: store.MountPath})
	}

	hostConfig := &container.HostConfig{
		PortBindings:  portBind,
		Mounts:        mounts,
		RestartPolicy: container.RestartPolicy{Name: "always"},
		CapAdd:        cont.CapAdd,
	}
	//EW HACKY
	if UseProxy {
		if cont.Name != "agogos-proxy" && cont.Name != "agogos-mongo-primary" && cont.Name != "agogos-mongo-secondary" &&
			cont.Name != "agogos-reverse-proxy" && cont.Name != "host.docker.internal" {
			hostConfig.Links = []string{"agogos-proxy"}
		}
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        cont.Image,
		Labels:       cont.Labels,
		Env:          cont.Env,
		ExposedPorts: ports,
	}, hostConfig, nil, cont.Name)
	if err != nil {
		return errors.New("Error Creating container: " + err.Error())
	}

	for _, net := range cont.Networks {
		settings := &network.EndpointSettings{}
		if cont.Alias != "" {
			settings.Aliases = []string{cont.Alias}
		}
		if err := cli.NetworkConnect(ctx, net.Name, resp.ID, settings); err != nil {
			return errors.New("error attaching container to network " + resp.ID + " " + err.Error())
		}
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return errors.New("Error starting container: " + err.Error())
	}

	return nil
}

func (DockerRuntime) ReadContainer(id string) (*Container, error) {
	cli := GetDockerClient()
	ctx := context.Background()

	filter := filters.NewArgs()
	filter.Add("name", id)
	cont, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filter,
		All:     true,
	})

	if err != nil {
		return nil, errors.New("Error listing containers: " + err.Error())
	}

	if len(cont) != 1 {
		return nil, errors.New("did not return single container for id " + id)
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
		return nil, errors.New("Error listing containers: " + err.Error())
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
		return nil, errors.New("Error listing containers: " + err.Error())
	}
	return dockerContainersToInterface(containers...)
}

func (DockerRuntime) UpdateContainer(container *Container) error {
	//TODO Implement
	panic("implement me")
}

func (DockerRuntime) DeleteContainer(id string) error {
	cli := GetDockerClient()
	err := cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Error deleting container %s: %s", id, err.Error()))
	}
	return nil
}

func (d DockerRuntime) DeleteContainerWithFilter(filter map[string]map[string]bool) error {
	cli := GetDockerClient()

	containers, err := d.ReadAllContainersWithFilter(filter)
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving containers with filter %v: %s", filter, err.Error()))
	}

	for _, cont := range containers {
		err = cli.ContainerRemove(context.Background(), cont.ID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
		if err != nil {
			return errors.New(fmt.Sprintf("Error deleting container %s: %s", cont.Name, err.Error()))
		}
	}
	return nil
}

func (DockerRuntime) CreateStorage(name string) error {
	cli := GetDockerClient()

	_, err := cli.VolumeCreate(context.Background(), volume.VolumeCreateBody{Name: name})
	if err != nil {
		return errors.New("Error creating storage: " + err.Error())
	}
	return nil
}

func (DockerRuntime) DeleteStorage(name string) error {
	cli := GetDockerClient()

	err := cli.VolumeRemove(context.Background(), name, false)
	if err != nil {
		return errors.New("Error deleting storage: " + err.Error())
	}
	return nil
}

func (d DockerRuntime) CreateNetwork(name string) error {
	cli := GetDockerClient()

	resp, err := cli.NetworkCreate(context.Background(), name, types.NetworkCreate{
		CheckDuplicate: true,
	})
	if err != nil {
		return errors.New("Error creating network: " + err.Error())
	}

	//Check for a network created with a bad subnet
	net, err := cli.NetworkInspect(context.Background(), resp.ID, types.NetworkInspectOptions{})
	if err != nil {
		return errors.New("Error checking network after creation: " + err.Error())
	}
	//Delete and recreate if the subnet is 192.168.0.0
	if net.IPAM.Config[0].Subnet == "192.168.0.0/20" {
		err := d.DeleteNetwork(name)
		if err != nil {
			return errors.New("error deleting invalid network: " + err.Error())
		}
		//Try to create the network again
		return d.CreateNetwork(name)
	}
	return nil
}

func (DockerRuntime) DeleteNetwork(name string) error {
	cli := GetDockerClient()

	err := cli.NetworkRemove(context.Background(), name)
	if err != nil {
		return errors.New("Error deleting network: " + err.Error())
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
		return false, errors.New("Error checking network: " + err.Error())
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
	stats, err := getStatsForContainer(dockerCont)
	if err != nil {
		return nil, err
	}
	if len(dockerCont.Names) < 0 {
		return nil, errors.New(fmt.Sprintf("Container %s has no name", dockerCont.ID))
	}
	stringPorts := make(map[string]string)
	for _, port := range dockerCont.Ports {
		stringPorts[strconv.Itoa(int(port.PublicPort))] = strconv.Itoa(int(port.PrivatePort))
	}

	var nets []Network
	for name, details := range dockerCont.NetworkSettings.Networks {
		//if name != "bridge" {
		nets = append(nets, Network{name, details.IPAddress})
		//}
	}

	var storage []StorageMount
	for _, contMount := range dockerCont.Mounts {
		storage = append(storage, StorageMount{Name: contMount.Name, MountPath: contMount.Destination})
	}

	//TODO try to get env here
	cont := &Container{
		ID:       dockerCont.ID,
		Image:    dockerCont.Image,
		Name:     dockerCont.Names[0],
		Labels:   dockerCont.Labels,
		Status:   dockerCont.State,
		Ports:    stringPorts,
		Storage:  storage,
		Stats:    *stats,
		Networks: nets,
	}
	return cont, nil
}

func getStatsForContainer(dockerCont types.Container) (*Stats, error) {
	cli := GetDockerClient()

	resp, err := cli.ContainerStats(context.Background(), dockerCont.ID, false)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var v *types.StatsJSON
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	previousCPU := v.PreCPUStats.CPUUsage.TotalUsage
	previousSystem := v.PreCPUStats.SystemUsage
	cpuPercent := calculateCPUPercentUnix(previousCPU, previousSystem, v)
	return &Stats{
		CpuUsage: cpuPercent,
	}, nil
}

func calculateCPUPercentUnix(previousCPU, previousSystem uint64, v *types.StatsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(previousSystem)
	)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(v.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}
