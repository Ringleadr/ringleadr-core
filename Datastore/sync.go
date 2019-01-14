package Datastore

import (
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Components"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"log"
	"strings"
	"time"
)

func startSync() {
	go func() {
		runtime := Containers.GetContainerRuntime()
		for {
			syncTick(runtime)
			time.Sleep(10 * time.Second)
		}
	}()
}

func syncTick(runtime Containers.ContainerRuntime) {
	//TODO Deal with networks and storage here too
	//TODO this doesn't create a network if an application needs one
	log.Printf("tick")
	//Check datastore is up
	cont, err := runtime.ReadContainer("agogos-mongo-primary")
	if !(err == nil && strings.Contains(cont.Status, "running")) {
		startDatastoreContainer(runtime)
		//might need to restart watchers here
	}
	//Get all items in the Applications database
	apps, err := GetAllApps()
	if err != nil {
		log.Printf("error getting Applications from datastore in sync thread: %s", err.Error())
		//We don't explicitly deal with the error here as we will come back around again in 10s and retry
	}

	containers, err := runtime.ReadAllContainers()
	if err != nil {
		log.Printf("error getting containers from runtime in sync thread: %s", err.Error())
		//We don't explicitly deal with the error here as we will come back around again in 10s and retry
	}

	//Look for components without matching containers (create missing)
	for _, app := range apps {
		for i := 0; i < app.Copies; i++ {
			for _, comp := range app.Components {
				if comp.Name == "" {
					comp.Name = comp.Image
				}
				for j := 0; j < comp.Replicas; j++ {
					if !lookForMatchingContainer("/"+Containers.GetContainerNameForComponent(comp.Name, app.Name, i, j), containers) {
						go Components.StartComponentReplica(comp, app.Name, i, app.Networks, j)
					}
				}
			}
		}
	}

	//Look for containers without matching components (and delete)
	for _, cont := range containers {
		//If the container is in created mode (and not running) then let's get rid of it.
		if cont.Status == "created: Created" {
			go runtime.DeleteContainer(cont.Name)
		}
		//If the app which owns the container no longer exists then purge
		if !lookForMatchingApplication(cont.Labels["agogos.owned.by"], apps) {
			go runtime.DeleteContainer(cont.Name)
		}
	}
}

func lookForMatchingContainer(containerName string, conts []*Containers.Container) bool {
	found := false
	for _, cont := range conts {
		if cont.Name == containerName {
			found = true
			break
		}
	}
	return found
}

func lookForMatchingApplication(applicationName string, apps []Datatypes.Application) bool {
	canonicalName := applicationName[:strings.LastIndex(applicationName, "-")]
	found := false
	for _, app := range apps {
		if app.Name == canonicalName {
			found = true
			break
		}
	}
	return found

}
