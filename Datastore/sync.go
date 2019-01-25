package Datastore

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Components"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/GodlikePenguin/agogos-host/Utils"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

func startSync(mode string, address string) {
	go func() {
		runtime := Containers.GetContainerRuntime()
		for {
			syncTick(runtime, mode, address)
			time.Sleep(1 * time.Second)
		}
	}()
}

func syncTick(runtime Containers.ContainerRuntime, mode string, address string) {
	//TODO remove old error messages if they are no longer valid
	hostname, err := os.Hostname()
	if err != nil {
		//skip this go and try again next time
		Logger.ErrPrintf("Error getting hostname in sync thread: %s", err.Error())
		return
	}
	//Check datastore is up
	if mode == "Primary" {
		cont, err := runtime.ReadContainer("agogos-mongo-primary")
		if !(err == nil && strings.Contains(cont.Status, "running")) {
			startDatastoreContainer(runtime)
		}
	} else if mode == "Secondary" {
		cont, err := runtime.ReadContainer("agogos-mongo-secondary")
		if !(err == nil && strings.Contains(cont.Status, "running")) {
			startSecondaryDatastoreContainer(runtime, address)
		}
	}
	//Get all items in the Applications database
	apps, err := GetAllApps()
	if err != nil {
		log.Printf("error getting Applications from datastore in sync thread: %s", err.Error())
		//We don't explicitly deal with the error here as we will come back around again in 10s and retry
		return
	}

	containers, err := runtime.ReadAllContainers()
	if err != nil {
		log.Printf("error getting containers from runtime in sync thread: %s", err.Error())
		//We don't explicitly deal with the error here as we will come back around again in 10s and retry
		return
	}

	networks, err := GetAllNetworks()
	if err != nil {
		log.Printf("error getting networks from datastore in sync thread: %s", err.Error())
		//We don't explicitly deal with the error here as we will come back around again in 10s and retry
		return
	}

	if mode == "Primary" {
		nodes, err := GetAllNodes()
		if err != nil {
			Logger.ErrPrintf("Error fetching nodes from sync thread: %s", err.Error())
		}
		//This is by far the slowest part of all the sync thread due to http timeouts on failure.
		// Consider calling this method less often?
		checkForInactiveNodes(nodes, hostname)
		distribution := getAppDistribution(apps, nodes)
		assignAppsWithNoNodes(&apps, distribution)
		displayErrorForAppsOnInactiveNodes(apps, nodes)
	}

	createMissingNetworks(networks, runtime)
	createMissingComponents(apps, containers, runtime, hostname)
	deleteOrphanedContainers(apps, containers, runtime, hostname)

}
func createMissingComponents(apps []Datatypes.Application, containers []*Containers.Container, runtime Containers.ContainerRuntime, hostname string) {
	//Look for components without matching containers (create missing)
	for _, app := range apps {
		if app.Node != hostname && app.Node != "*" {
			continue
		}
		shouldSave := false
		for i := 0; i < app.Copies; i++ {
			for _, comp := range app.Components {
				var compCPUTotal float64
				if comp.Name == "" {
					comp.Name = comp.Image
				}
				for j := 0; j < comp.Replicas; j++ {
					if matchingCont := lookForMatchingContainer("/"+Containers.GetContainerNameForComponent(comp.Name, app.Name, i, j), containers); matchingCont == nil {
						createMissingAppNetworks(app, runtime)
						createMissingStorage(comp, runtime)
						err := Components.StartComponentReplica(comp, app.Name, i, app.Networks, j)
						if err != nil {
							Logger.ErrPrintf("Error starting component %s in app %s: %s", comp.Name, app.Name, err.Error())
							formatError := fmt.Sprintf("Error starting component %s: %s", comp.Name, err.Error())
							if !Utils.StringArrayContains(app.Messages, formatError) {
								app.Messages = append(app.Messages, formatError)
								shouldSave = true
							}
						}
					} else {
						if matchingCont.Status != comp.Status {
							comp.Status = matchingCont.Status
							shouldSave = true
						}
						compCPUTotal += matchingCont.Stats.CpuUsage
					}
				}
				if comp.ScaleThreshold != 0 {
					desiredReplicas := math.Ceil(float64(comp.Replicas) * ((compCPUTotal / float64(comp.Replicas)) / float64(comp.ScaleThreshold)))
					if int(desiredReplicas) > comp.ScaleMax {
						desiredReplicas = float64(comp.ScaleMax)
					} else if int(desiredReplicas) < comp.ScaleMin {
						desiredReplicas = float64(comp.ScaleMin)
					}
					if int(desiredReplicas) != comp.Replicas {
						comp.Replicas = int(desiredReplicas)
						shouldSave = true
					}

				}
			}
		}
		if shouldSave {
			err := UpdateApp(&app)
			if err != nil {
				Logger.ErrPrintf("Error updating %s in application datastore: %s", app.Name, err.Error())
			}
		}
	}
}

func createMissingAppNetworks(app Datatypes.Application, runtime Containers.ContainerRuntime) {
	for _, net := range app.Networks {
		formatName := fmt.Sprintf("agogos-%s", net)
		fetchedNetwork, err := GetNetwork(formatName)
		if err != nil {
			Logger.ErrPrintf("Error fetching network %s from datastore: %s", formatName, err.Error())
			//continue and hope its fixed next go round
			continue
		}
		if fetchedNetwork == nil {
			err = InsertNetwork(&Datatypes.Network{Name: formatName})
			if err != nil {
				Logger.Printf("Error inserting network %s: %s", formatName, err.Error())
			}
		}
	}
	//Give a little bit of time for the networks to start
	time.Sleep(1 * time.Second)
	//Create implicit network for this app
	exists, err := runtime.NetworkExists(app.Name)
	if err != nil {
		Logger.ErrPrintf("Error checking for existing implicit network %s: %s", app.Name, err.Error())
		//Hope it's fixed next go round
	} else {
		if !exists {
			err = Containers.GetContainerRuntime().CreateNetwork(app.Name)
			if err != nil {
				Logger.ErrPrintf("Error creating implicit network for %s: %s", app.Name, err.Error())
			}
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

func lookForMatchingContainer(containerName string, conts []*Containers.Container) *Containers.Container {
	for _, cont := range conts {
		if cont.Name == containerName {
			return cont
		}
	}
	return nil
}

func deleteOrphanedContainers(apps []Datatypes.Application, containers []*Containers.Container, runtime Containers.ContainerRuntime, hostname string) {
	//Look for containers without matching components (and delete)
	for _, cont := range containers {
		//If the container is in created mode (and not running) then let's get rid of it.
		if cont.Status == "created" {
			go func(name string) {
				err := runtime.DeleteContainer(name)
				if err != nil {
					Logger.ErrPrintf("Error deleting container stuck in creating: %s", name, err.Error())
				}
			}(cont.Name)
		}
		//If the app which owns the container no longer exists then purge
		if !lookForMatchingApplication(cont.Labels["agogos.owned.by"], apps, hostname) {
			go func(name string) {
				err := runtime.DeleteContainer(name)
				if err != nil {
					//Don't repeat the name of the container here, the underlying error already contains the name
					Logger.ErrPrintln(err.Error())
				}
			}(cont.Name)
		}
	}
}

func lookForMatchingApplication(applicationName string, apps []Datatypes.Application, hostname string) bool {
	canonicalName := applicationName[:strings.LastIndex(applicationName, "-")]
	found := false
	for _, app := range apps {
		if app.Name == canonicalName && (app.Node == hostname || app.Node == "*") {
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
			go func(name string) {
				err := runtime.CreateNetwork(name)
				if err != nil {
					Logger.ErrPrintf("Error creating network %s: %s", name, err.Error())
				}
			}(net.Name)
		}
	}
}

func getAppDistribution(apps []Datatypes.Application, nodes []Datatypes.Node) map[string]int {
	results := make(map[string]int)
	for _, node := range nodes {
		if node.Active {
			results[node.Name] = 0
		}
	}
	for _, app := range apps {
		if app.Node != "" && app.Node != "*" {
			results[app.Node] = results[app.Node] + 1
		}
	}
	for _, node := range nodes {
		if !node.Active {
			//Remove any inactive nodes from the results so we don't schedule a container there
			delete(results, node.Name)
		}
	}
	return results
}

func assignAppsWithNoNodes(apps *[]Datatypes.Application, distribution map[string]int) {
	for _, app := range *apps {
		if app.Node != "" {
			continue
		}
		selectedNode := Utils.GetMinFromStringIntMap(distribution)
		if selectedNode != "" {
			app.Node = selectedNode
			distribution[selectedNode] = distribution[selectedNode] + 1
			err := UpdateApp(&app)
			if err != nil {
				Logger.ErrPrintf("Error assigning node to app: %s", err.Error())
			}
		}
	}
}

func checkForInactiveNodes(nodes []Datatypes.Node, hostname string) {
	for _, node := range nodes {
		if node.Name == hostname {
			continue
		}
		oldStatus := node.Active
		resp, err := http.Get(fmt.Sprintf("http://%s:14440/ping", node.Address))
		if err != nil || resp.StatusCode != http.StatusOK {
			node.Active = false
		} else {
			node.Active = true
		}
		if oldStatus != node.Active {
			err := UpdateNode(&node)
			if err != nil {
				Logger.ErrPrintf("Error updating node status in sync thread: %s", err.Error())
			}
		}
	}
}

func displayErrorForAppsOnInactiveNodes(apps []Datatypes.Application, nodes []Datatypes.Node) {
	for _, app := range apps {
		shouldSave := false
		for _, node := range nodes {
			if app.Node == node.Name {
				if !node.Active {
					formatError := fmt.Sprintf("Application is scheduled on inactive node: %s", node.Name)
					if !Utils.StringArrayContains(app.Messages, formatError) {
						app.Messages = append(app.Messages, formatError)
						shouldSave = true
					}
				}
			}
		}
		if shouldSave {
			err := UpdateApp(&app)
			if err != nil {
				Logger.ErrPrintf("Error saving inactive node message to application: %s", err.Error())
			}
		}
	}
}
