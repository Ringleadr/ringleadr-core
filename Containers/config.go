package Containers

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/GodlikePenguin/agogos-host/Utils"
	"github.com/docker/docker/client"
	"os"
	"runtime"
	"strings"
)

var dockerClient *client.Client
var containerRuntime ContainerRuntime
var UseProxy bool

func SetupConfig(contRuntime ContainerRuntime, useProxy bool) {
	UseProxy = useProxy
	cli, err := client.NewClientWithOpts(client.WithVersion("1.39"))
	if err != nil {
		panic(err)
	}
	dockerClient = cli
	containerRuntime = contRuntime
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

func StartProxies() {
	if UseProxy {
		exists, err := containerRuntime.NetworkExists("agogos-proxy-net")
		if err != nil {
			panic("could not check for existing agogos-proxy network: " + err.Error())
		}
		if !exists {
			err = containerRuntime.CreateNetwork("agogos-proxy-net")
			if err != nil {
				panic("Error setting up proxy network: " + err.Error())
			}
		}
		//Docker on linux does not have a DNS entry for for 'host.docker.internal' so we fake one with a container
		if runtime.GOOS == "linux" {
			startDockerHostProxy()
		}
		startProxy()
		startReverseProxy()
	}
}

func startDockerHostProxy() {
	cont, err := containerRuntime.ReadContainer("host.docker.internal")
	if err != nil {
		if strings.Contains(err.Error(), "did not return single container") {
			//Doesn't exist, let's make a new one
		} else {
			panic("Error checking for existing host proxy")
		}
	} else if strings.Contains(cont.Status, "running") {
		Logger.Println("Using existing host proxy")
		return
	}

	Logger.Println("Starting host proxy")

	hostProxy := &Container{
		Name:  "host.docker.internal",
		Image: "edwarddobson/docker-host:" + Utils.GetEnvOrDefault("AGOGOS_HOST_PROXY_TAG", "latest"),
		Labels: map[string]string{
			"agogos-host-proxy": "",
		},
		Alias: "host.docker.internal",
		Networks: []Network{
			{Name: "agogos-proxy-net"},
		},
		CapAdd: []string{
			"NET_ADMIN",
			"NET_RAW",
		},
	}
	err = containerRuntime.CreateContainer(hostProxy)
	if err != nil {
		panic("Unable to start Agogos host proxy: " + err.Error())
	}
}

func startProxy() {
	hostname, err := os.Hostname()
	if err != nil {
		panic("Can't get hostname to pass to agogos proxy. Exiting.")
	}

	cont, err := containerRuntime.ReadContainer("agogos-proxy")
	if err != nil {
		if strings.Contains(err.Error(), "did not return single container") {
			//Doesn't exist, let's make a new one
		} else {
			panic("Error checking for existing host proxy")
		}
	} else if strings.Contains(cont.Status, "running") {
		Logger.Println("Using existing host proxy")
		return
	}

	Logger.Println("Starting proxy")

	proxy := &Container{
		Name:  "agogos-proxy",
		Image: "edwarddobson/agogos-proxy:" + Utils.GetEnvOrDefault("AGOGOS_PROXY_TAG", "latest"),
		Networks: []Network{
			{Name: "agogos-proxy-net"},
		},
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

func startReverseProxy() {
	hostname, err := os.Hostname()
	if err != nil {
		panic("Can't get hostname to pass to agogos reverse proxy. Exiting.")
	}

	cont, err := containerRuntime.ReadContainer("agogos-reverse-proxy")
	if err != nil {
		if strings.Contains(err.Error(), "did not return single container") {
			//Doesn't exist, let's make a new one
		} else {
			panic("Error checking for existing host proxy")
		}
	} else if strings.Contains(cont.Status, "running") {
		Logger.Println("Using existing host proxy")
		return
	}

	Logger.Println("Starting reverse proxy")

	proxy := &Container{
		Name:  "agogos-reverse-proxy",
		Image: "edwarddobson/agogos-reverse-proxy:" + Utils.GetEnvOrDefault("AGOGOS_REVERSE_PROXY_TAG", "latest"),
		Networks: []Network{
			{Name: "agogos-proxy-net"},
		},
		Labels: map[string]string{
			"agogos-reverse-proxy": "",
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
