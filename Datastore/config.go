package Datastore

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/GodlikePenguin/agogos-host/Utils"
	"github.com/globalsign/mgo"
	"os"
	"strings"
	"time"
)

var (
	mongoClient           *mgo.Session
	agogosDB              *mgo.Database
	applicationCollection *mgo.Collection
	componentCollection   *mgo.Collection
	storageCollection     *mgo.Collection
	networkCollection     *mgo.Collection
	nodesCollection       *mgo.Collection
)

func SetupDatastore(mode string, primaryAddress string) {
	//Container runtime should be set up by now, and we know it's running.
	//Let's create a Datastore container
	runtime := Containers.GetContainerRuntime()

	if mode == "Primary" {
		//Only create the container if one doesn't exist
		cont, err := runtime.ReadContainer("agogos-mongo-primary")
		if !(err == nil && strings.Contains(cont.Status, "running")) {

			startDatastoreContainer(runtime)

			//Sleep to give time for db to start
			//TODO do this is a more programatic way
			time.Sleep(1 * time.Minute)
		} else {
			Logger.Println("Using existing database")
		}
	} else if mode == "Secondary" {
		//Only create the container if one doesn't exist
		/*
			cont, err := runtime.ReadContainer("agogos-mongo-secondary")
			if !(err == nil && strings.Contains(cont.Status, "running")) {

				startSecondaryDatastoreContainer(runtime, primaryAddress)

				//Sleep to give time for db to start
				//TODO do this is a more programatic way
				time.Sleep(30 * time.Second)
			} else {
				Logger.Println("Using existing database")
			}
		*/
	}

	getClient(strings.ToLower(mode), primaryAddress)
	setupTables()
	if mode == "Primary" {
		addThisNode()
	}
	//startWatchers()
	startSync(mode, primaryAddress)
}

func startDatastoreContainer(runtime Containers.ContainerRuntime) {
	Logger.Println("Creating new datastore. This may take some time")

	//Create a new data store
	config := &Containers.Container{
		Name:  "agogos-mongo-primary",
		Image: "bitnami/mongodb:3.6.8",
		Labels: map[string]string{
			"agogos-mongo": "primary",
		},
		Env: []string{
			"MONGODB_REPLICA_SET_MODE=primary",
		},
		Ports: map[string]string{
			"0.0.0.0:27017": "27017",
		},
		Storage: []Containers.StorageMount{{Name: "agogos-mongo-primary-storage", MountPath: "/bitnami"}},
	}

	if err := runtime.CreateContainer(config); err != nil {
		panic(fmt.Sprintf("Could not start backing Datastore: %s", err.Error()))
	}
}

func startSecondaryDatastoreContainer(runtime Containers.ContainerRuntime, address string) {
	Logger.Println("Creating new datastore. This may take some time.")

	//Create a new data store
	config := &Containers.Container{
		Name:  "agogos-mongo-secondary",
		Image: "bitnami/mongodb:3.6.8",
		Labels: map[string]string{
			"agogos-mongo": "secondary",
		},
		Env: []string{
			"MONGODB_REPLICA_SET_MODE=secondary",
			fmt.Sprintf("MONGODB_PRIMARY_HOST=%s", address),
		},
		Ports: map[string]string{
			"0.0.0.0:27017": "27017",
		},
		Storage: []Containers.StorageMount{{Name: "agogos-mongo-secondary-storage", MountPath: "/bitnami"}},
	}

	if err := runtime.CreateContainer(config); err != nil {
		panic(fmt.Sprintf("Could not start backing Datastore: %s", err.Error()))
	}
}

func getClient(mode string, address string) {
	//Wait until the service is ready
	if mode == "primary" {
		waitUntilReady(mode)
	}
	//setup the client
	mongoClient = setupClient(address)
}

func setupTables() {
	//create the db and the collection
	agogosDB = mongoClient.DB("agogos")

	//Hacky insert to ensure the DB exists for the watcher
	dummy := map[string]string{
		"foo": "bah",
	}
	err := agogosDB.C("foo").Insert(dummy)
	if err != nil {
		panic(err)
	}

	applicationCollection = agogosDB.C("applications")
	componentCollection = agogosDB.C("components")
	storageCollection = agogosDB.C("storage")
	networkCollection = agogosDB.C("networks")
	nodesCollection = agogosDB.C("nodes")
}

func startWatchers() {
	go watchApplications(applicationCollection)
	go watchStorage(storageCollection)
	go watchNetworks(networkCollection)
}

func waitUntilReady(mode string) {
	runtime := Containers.GetContainerRuntime()
	var lastErr error
	for i := 0; i < 20; i++ {
		cont, err := runtime.ReadContainer(fmt.Sprintf("agogos-mongo-%s", mode))
		time.Sleep(2 * time.Second)
		if err != nil {
			lastErr = err
			continue
		}
		if strings.Contains(cont.Status, "running") {
			return
		}
		cont, err = runtime.ReadContainer(fmt.Sprintf("agogos-mongo-%s", mode))
	}
	panic("Could not check container status after 20 attempts. Quitting. Last error was: " + lastErr.Error())
}

func setupClient(address string) *mgo.Session {
	if address == "" {
		address = "localhost"
	} else {
		Logger.Printf("Connecting to primary datastore at %s", address)
	}
	session, err := mgo.Dial(fmt.Sprintf("mongodb://%s:27017", address))
	if err != nil {
		panic(err)
	}
	return session
}

func addThisNode() {
	name, err := os.Hostname()
	if err != nil {
		panic("Could not get hostname: " + err.Error())
	}
	address := Utils.GetOutboundIP()
	node, err := GetNode(name)
	if err != nil {
		panic("Could not check node collection on startup")
	}
	if node == nil {
		err = InsertNode(&Datatypes.Node{Name: name, Address: address.String(), Active: true})
		if err != nil {
			panic("Could not set up datastore with this nodes information: " + err.Error())
		}
	}
}
