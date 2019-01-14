package Datastore

import (
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

	for _, app := range apps {
		for i := 0; i < app.Copies; i++ {
			for _, comp := range app.Components {
				if comp.Name == "" {
					comp.Name = comp.Image
				}
				for j := 0; j < comp.Replicas; j++ {
					if !checkFor("/"+Containers.GetContainerNameForComponent(comp.Name, app.Name, i, j), containers) {
						go Components.StartComponentReplica(comp, app.Name, i, app.Networks, j)
					}
				}
			}
		}
	}
}

func checkFor(containerName string, conts []*Containers.Container) bool {
	found := false
	for _, cont := range conts {
		if cont.Name == containerName {
			found = true
			break
		}
	}
	return found
}
