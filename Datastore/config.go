package Datastore

import (
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/globalsign/mgo"
	"strings"
	"time"
)

var (
	mongoClient           *mgo.Session
	agogosDB              *mgo.Database
	applicationCollection *mgo.Collection
	storageCollection     *mgo.Collection
	networkCollection     *mgo.Collection
)

func SetupDatastore() {
	//Container runtime should be set up by now, and we know it's running.
	//Let's create a Datastore container
	runtime := Containers.GetContainerRuntime()

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

	getClient()
	setupTables()
	//startWatchers()
	startSync()
}

func startDatastoreContainer(runtime Containers.ContainerRuntime) {
	Logger.Println("Creating new datastore. This may take some time.")
	//TODO give a docker volume

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
			"27017": "27017",
		},
		Storage: []Containers.StorageMount{{Name: "agogos-mongo-primary-storage", MountPath: "/bitnami"}},
	}

	if err := runtime.CreateContainer(config); err != nil {
		panic("Could not start backing Datastore")
	}
}

func getClient() {
	//Wait until the service is ready
	waitUntilReady()
	//setup the client
	mongoClient = setupClient()
}

func setupTables() {
	//create the db and the collection
	db := mongoClient.DB("agogos")
	agogosDB = db

	//Hacky insert to ensure the DB exists for the watcher
	dummy := map[string]string{
		"foo": "bah",
	}
	err := db.C("foo").Insert(dummy)
	if err != nil {
		panic(err)
	}

	coll := db.C("applications")
	applicationCollection = coll
	storage := db.C("storage")
	storageCollection = storage
	network := db.C("networks")
	networkCollection = network
}

func startWatchers() {
	go watchApplications(applicationCollection)
	go watchStorage(storageCollection)
	go watchNetworks(networkCollection)
}

func waitUntilReady() {
	runtime := Containers.GetContainerRuntime()
	cont, err := runtime.ReadContainer("agogos-mongo-primary")
	for !strings.Contains(cont.Status, "running") {
		if err != nil {
			panic("Could not set up database")
		}
		time.Sleep(2 * time.Second)
		cont, err = runtime.ReadContainer("agogos-mongo-primary")
	}
}

func setupClient() *mgo.Session {
	session, err := mgo.Dial("mongodb://localhost:27017")
	if err != nil {
		panic(err)
	}
	return session
}
