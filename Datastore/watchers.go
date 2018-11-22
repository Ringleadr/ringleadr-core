package Datastore

import (
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Components"
	"github.com/davecgh/go-spew/spew"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"log"
)

//Watch actions

func watchApplications(coll *mgo.Collection) {
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
			var app Datatypes.Application
			if changeDoc["operationType"] == "insert" {
				unMarshalIntoApp(changeDoc, &app)
				createComponentsFor(&app)
			} else if changeDoc["operationType"] == "delete" {
				//On a delete event, we can't get the full data or even just the name from mongo, so here we
				// don't use a changestream, and rely on the handler calling the DeleteAllComponentsForApp method
			}
			//TODO rest
			spew.Dump(changeDoc)
		}
	}
}

func unMarshalIntoApp(m bson.M, app *Datatypes.Application) {
	bsonBytes, err := bson.Marshal(m["fullDocument"])
	if err != nil {
		log.Println(err)
		return
	}
	bson.Unmarshal(bsonBytes, &app)
}

func createComponentsFor(app *Datatypes.Application) {
	for _, comp := range app.Components {
		go Components.StartComponent(comp, app.Name)
	}
}
