package Containers

import (
	"fmt"
	"github.com/docker/docker/client"
	"strings"
)

var dockerClient *client.Client
var containerRuntime ContainerRuntime

func SetupConfig(runtime ContainerRuntime) {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.39"))
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

func GetContainerNameForComponent(componentName string, appName string, appCopy int, replicaNo int) string {
	return strings.Replace(
		strings.Replace(
			fmt.Sprintf("agogos-%s-%d-%s-%d", appName, appCopy, componentName, replicaNo),
			":", "_", -1), "/", ".", -1)
}
