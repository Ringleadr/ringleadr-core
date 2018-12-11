package Containers

import (
	"fmt"
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
	err = containerRuntime.AssertOnline()
	if err != nil {
		panic(err)
	}
}

func GetDockerClient() *client.Client {
	return dockerClient
}

func GetContainerRuntime() ContainerRuntime {
	return containerRuntime
}

func GetContainerNameForComponent(componentName string, appName string, replicaNo int) string {
	return fmt.Sprintf("agogos-%s-%s-%d", appName, componentName, replicaNo)
}
