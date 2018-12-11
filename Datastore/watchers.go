package Datastore

import (
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Components"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"log"
)

//Watch actions

func watchApplications(coll *mgo.Collection) {
	var insertFunc = func(changeDoc bson.M) {
		var app Datatypes.Application
		unMarshalIntoApp(changeDoc, &app)
		createComponentsFor(&app)
	}

	var deleteFunc = func(changeDoc bson.M) {
		//On a delete event, we can't get the full data or even just the name from mongo, so here we
		// don't use a changestream, and rely on the handler calling the DeleteAllComponentsForApp method
	}

	var funcs = map[string]func(m bson.M){
		"insert": insertFunc,
		"delete": deleteFunc,
	}

	watchGeneralCollection(coll, funcs)
}

func watchStorage(coll *mgo.Collection) {
	var insertFunc = func(changeDoc bson.M) {
		var storage Datatypes.Storage
		unMarshalIntoStorage(changeDoc, &storage)
		err := CreateStorageInRuntime(storage.Name)
		if err != nil {
			log.Println(err)
		}
	}

	var deleteFunc = func(changeDoc bson.M) {
		//Again, mongo delete events don't contain the data. no-op function
	}

	var funcs = map[string]func(m bson.M){
		"insert": insertFunc,
		"delete": deleteFunc,
	}

	watchGeneralCollection(coll, funcs)
}

func watchGeneralCollection(coll *mgo.Collection, funcs map[string]func(changeDoc bson.M)) {
	var pipeline []bson.M
	var changeDoc bson.M

	stream, err := coll.Watch(pipeline, mgo.ChangeStreamOptions{})
	defer stream.Close()
	if err != nil {
		panic(err)
	}
	//Infinite for loop to keep watching the stream (even though Next should block?)
	for {
		for stream.Next(&changeDoc) {
			if changeDoc["operationType"] == "insert" {
				funcs["insert"](changeDoc)
			} else if changeDoc["operationType"] == "delete" {
				funcs["delete"](changeDoc)
			}
			//TODO rest
			//spew.Dump(changeDoc)
		}
	}
}

func unMarshalIntoApp(m bson.M, app *Datatypes.Application) {
	bsonBytes, err := bson.Marshal(m["fullDocument"])
	if err != nil {
		log.Println(err)
		return
	}
	err = bson.Unmarshal(bsonBytes, &app)
	if err != nil {
		log.Print(err)
	}
}

func unMarshalIntoStorage(m bson.M, app *Datatypes.Storage) {
	bsonBytes, err := bson.Marshal(m["fullDocument"])
	if err != nil {
		log.Println(err)
		return
	}
	err = bson.Unmarshal(bsonBytes, &app)
	if err != nil {
		log.Print(err)
	}
}

func createComponentsFor(app *Datatypes.Application) {
	for _, comp := range app.Components {
		go Components.StartComponent(comp, app.Name)
	}
}

///////////////////// THESE METHODS DO NOT BELONG IN THIS PACKAGE, BUT ARE HERE TO STOP IMPORT CYCLES. REDESIGN NEEDED //

func CreateStorageInRuntime(name string) error {
	runtime := Containers.GetContainerRuntime()
	err := runtime.CreateStorage(name)
	if err != nil {
		return err
	}
	return nil
}

func DeleteStorageInRuntime(name string) error {
	runtime := Containers.GetContainerRuntime()
	err := runtime.DeleteStorage(name)
	if err != nil {
		return err
	}
	return nil
}
