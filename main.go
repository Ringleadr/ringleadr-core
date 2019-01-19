package main

import (
	"github.com/GodlikePenguin/agogos-host/Applications"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Datastore"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/GodlikePenguin/agogos-host/Networks"
	"github.com/GodlikePenguin/agogos-host/Storage"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"runtime"
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
	"/networks":          Networks.ListNetworks,
}

var postMethods = map[string]func(ctx *gin.Context){
	"/applications":   Applications.CreateApplication,
	"/storage/:name":  Storage.CreateStorage,
	"/networks/:name": Networks.CreateNetwork,
}

var deleteMethods = map[string]func(ctx *gin.Context){
	"/applications/:name": Applications.DeleteApplication,
	"/storage/:name":      Storage.DeleteStorage,
	"/networks/:name":     Networks.DeleteNetwork,
}

var putMethods = map[string]func(ctx *gin.Context){
	"/applications": Applications.UpdateApplication,
}

func main() {
	//TODO be able to set from command line
	agogosMode := "Primary"
	Logger.Printf("Starting Agogos in %s mode", agogosMode)
	//TODO Take from environment
	containerRuntime := Containers.DockerRuntime{}
	Containers.SetupConfig(containerRuntime)

	//Use multiple cores for efficiency
	runtime.GOMAXPROCS(int(math.Min(float64(runtime.NumCPU()), 4)))

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
	for path, handler := range putMethods {
		r.PUT(path, handler)
	}

	log.Fatal(r.Run(":14440"))
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	return gin.Default()
}
