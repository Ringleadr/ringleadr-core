package main

import (
	"github.com/GodlikePenguin/agogos-host/Applications"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/gin-gonic/gin"
)

var getMethods = map[string]func(ctx *gin.Context){
	"/ping": func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	},
	"/applications": Applications.GetApplications,
}

var postMethods = map[string]func(ctx *gin.Context){
	"/applications": Applications.CreateApplication,
}

func main() {
	r := setupRouter()
	for path, handler := range getMethods {
		r.GET(path, handler)
	}
	for path, handler := range postMethods {
		r.POST(path, handler)
	}

	//TODO Take from environment
	runtime := Containers.DockerRuntime{}
	Containers.SetupConfig(runtime)
	r.Run(":14440")
}

func setupRouter() *gin.Engine {
	return gin.Default()
}
