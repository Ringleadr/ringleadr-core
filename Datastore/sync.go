package Datastore

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Components"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Logger"
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
	//TODO this doesn't create a network if an application needs one
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

	networks, err := GetAllNetworks()
	if err != nil {
		log.Printf("error getting networks from datastore in sync thread: %s", err.Error())
		//We don't explicitly deal with the error here as we will come back around again in 10s and retry
	}

	createMissingNetworks(networks, runtime)
	createMissingComponents(apps, containers, runtime)
	deleteOrphanedContainers(apps, containers, runtime)

}
func createMissingComponents(apps []Datatypes.Application, containers []*Containers.Container, runtime Containers.ContainerRuntime) {
	//Look for components without matching containers (create missing)
	for _, app := range apps {
		for i := 0; i < app.Copies; i++ {
			for _, comp := range app.Components {
				if comp.Name == "" {
					comp.Name = comp.Image
				}
				for j := 0; j < comp.Replicas; j++ {
					if !lookForMatchingContainer("/"+Containers.GetContainerNameForComponent(comp.Name, app.Name, i, j), containers) {
						createMissingAppNetworks(app, runtime)
						createMissingStorage(comp, runtime)
						err := Components.StartComponentReplica(comp, app.Name, i, app.Networks, j)
						if err != nil {
							Logger.ErrPrintf("Error starting component %s in app %s: %s", comp.Name, app.Name, err.Error())
							formatError := fmt.Sprintf("Error starting component %s: %s", comp.Name, err.Error())
							if !stringArrayContains(app.Messages, formatError) {
								app.Messages = append(app.Messages, formatError)
								err := UpdateApp(&app)
								if err != nil {
									Logger.ErrPrintln("Error saving error message to the application datastore: ", err.Error())
								}
							}
						}
					}
				}
			}
		}
	}
}

func stringArrayContains(arr []string, element string) bool {
	for _, a := range arr {
		if a == element {
			return true
		}
	}
	return false
}

func createMissingAppNetworks(app Datatypes.Application, runtime Containers.ContainerRuntime) {
	for _, net := range app.Networks {
		formatName := fmt.Sprintf("agogos-%s", net)
		fetchedNetwork, err := GetNetwork(formatName)
		if err != nil {
			//TODO something with error
			continue
		}
		if fetchedNetwork == nil {
			_ = InsertNetwork(&Datatypes.Network{Name: formatName})
		}
	}
	//Give a little bit of time for the networks to start
	time.Sleep(1 * time.Second)
	//Create implicit network for this app
	exists, err := runtime.NetworkExists(app.Name)
	if err != nil {
		//TODO deal with err better
	} else {
		if !exists {
			_ = Containers.GetContainerRuntime().CreateNetwork(app.Name)
		}
	}
}

func createMissingStorage(comp *Datatypes.Component, runtime Containers.ContainerRuntime) {
	for _, store := range comp.Storage {
		if s, err := GetStorage(fmt.Sprintf("agogos-%s", store.Name)); s == nil && err == nil {
			err := InsertStorage(&Datatypes.Storage{Name: fmt.Sprintf("agogos-%s", store.Name)})
			if err != nil {
				continue //skip this storage, should be cleaned up later by a watcher
			}
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

func deleteOrphanedContainers(apps []Datatypes.Application, containers []*Containers.Container, runtime Containers.ContainerRuntime) {
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

func createMissingNetworks(networks []Datatypes.Network, runtime Containers.ContainerRuntime) {
	for _, net := range networks {
		exists, err := runtime.NetworkExists(net.Name)
		if err != nil {
			log.Printf("error checking network %s in runtime from sync thread. Error was: %s", net.Name, err.Error())
			continue
		}
		if !exists {
			go runtime.CreateNetwork(net.Name)
		}
	}
}
