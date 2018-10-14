package Containers

import (
	"github.com/docker/docker/client"
)

var dockerClient *client.Client
var containerRuntime ContainerRuntime

func SetupConfig(runtime ContainerRuntime) {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	dockerClient = cli
	containerRuntime = runtime
}

func GetDockerClient() *client.Client {
	return dockerClient
}

func GetContainerRuntime() ContainerRuntime {
	return containerRuntime
}
