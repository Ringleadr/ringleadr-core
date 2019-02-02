package Containers

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/docker/docker/client"
	"os"
	"strings"
)

var dockerClient *client.Client
var containerRuntime ContainerRuntime
var UseProxy bool

func SetupConfig(runtime ContainerRuntime, useProxy bool) {
	UseProxy = useProxy
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
	if UseProxy {
		startProxy()
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

func startProxy() {
	hostname, err := os.Hostname()
	if err != nil {
		panic("Can't get hostname to pass to agogos proxy. Exiting.")
	}
	Logger.Println("Starting proxy")
	//Delete existing proxy if it exists:
	_ = containerRuntime.DeleteContainer("agogos-proxy")

	proxy := &Container{
		Name:  "agogos-proxy",
		Image: "edwarddobson/agogos-proxy",
		Labels: map[string]string{
			"agogos-proxy": "",
		},
		Env: []string{
			"AGOGOS_HOSTNAME=" + hostname,
		},
	}
	err = containerRuntime.CreateContainer(proxy)
	if err != nil {
		panic("Unable to start Agogos proxy: " + err.Error())
	}
}
