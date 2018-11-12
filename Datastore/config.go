package Datastore

import (
	"context"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/mongodb/mongo-go-driver/mongo"
	"log"
	"strings"
	"time"
)

var (
	mongoClient *mongo.Client
)

func SetupDatastore() {
	//Container runtime should be set up by now, and we know it's running.
	//Let's create a Datastore container
	runtime := Containers.GetContainerRuntime()

	//Check there isn't an existing datastore
	cont, err := runtime.ReadContainer("agogos-mongo-primary")
	if err == nil && strings.Contains(cont.Status, "running") {
		//already have a container, don't mess with it
		return
	}

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
	}

	if err := runtime.CreateContainer(config); err != nil {
		panic("Could not start backing Datastore")
	}

	go setupTables()
}

func setupTables() {
	//Wait until the service is ready
	waitUntilReady()
	//setup the client
	mongoClient = setupClient()
	//create the db and the collection
	db := mongoClient.Database("agogos")
	_ = db.Collection("applications")
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

func setupClient() *mongo.Client {
	client, err := mongo.NewClient("mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)

	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return client
}
