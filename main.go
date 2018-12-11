package main

import (
	"github.com/GodlikePenguin/agogos-host/Applications"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Datastore"
	"github.com/GodlikePenguin/agogos-host/Storage"
	"github.com/gin-gonic/gin"
	"log"
)

var getMethods = map[string]func(ctx *gin.Context){
	"/ping": func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	},
	"/applications":      Applications.GetApplications,
	"/application/:name": Applications.GetApplication,
	"/storage":           Storage.ListStorage,
}

var postMethods = map[string]func(ctx *gin.Context){
	"/applications":  Applications.CreateApplication,
	"/storage/:name": Storage.CreateStorage,
}

var deleteMethods = map[string]func(ctx *gin.Context){
	"/applications/:name": Applications.DeleteApplication,
	"/storage/:name":      Storage.DeleteStorage,
}

func main() {
	//TODO Take from environment
	runtime := Containers.DockerRuntime{}
	Containers.SetupConfig(runtime)

	Datastore.SetupDatastore()
	r := setupRouter()
	for path, handler := range getMethods {
		r.GET(path, handler)
	}
	for path, handler := range postMethods {
		r.POST(path, handler)
	}
	for path, handler := range deleteMethods {
		r.DELETE(path, handler)
	}

	log.Fatal(r.Run(":14440"))
}

func setupRouter() *gin.Engine {
	return gin.Default()
}
