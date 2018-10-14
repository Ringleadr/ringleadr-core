package Applications

import (
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Datatypes"
	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
)

//Create new application
func CreateApplication(ctx *gin.Context) {
	//Parse response into variable
	app := &Datatypes.Application{}
	ctx.BindJSON(app)
	spew.Dump(app)
	//Save state to DB
	//TODO

	//Start all components
	//TODO
}

func GetApplications(ctx *gin.Context) {
	//List all running applications
	//This will look at the DB and not Docker

	//For now return all agogos managed containers
	r := Containers.GetContainerRuntime()
	containers, err := r.ReadAllContainers()
	if err != nil {
		ctx.JSON(500, err)
	} else {
		ctx.JSON(200, containers)
	}
}

func GetApplication() {
	//Get a specific application
}

func UpdateApplication() {
	//Update a specific application
}

func DeleteApplication() {
	//Delete an application
}
