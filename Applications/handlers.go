package Applications

import (
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Components"
	"github.com/GodlikePenguin/agogos-host/Datastore"
	"github.com/gin-gonic/gin"
	"net/http"
)

//Create new application
func CreateApplication(ctx *gin.Context) {
	//Parse response into variable
	app := &Datatypes.Application{}
	ctx.BindJSON(app)

	err := Datastore.InsertApp(app)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func GetApplications(ctx *gin.Context) {
	apps, err := Datastore.GetAllApps()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, apps)
}

func GetApplication() {
	//Get a specific application
}

func UpdateApplication() {
	//Update a specific application
}

func DeleteApplication(ctx *gin.Context) {
	name := ctx.Param("name")
	err := Datastore.DeleteApp(name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}
	//Changestreams don't handle deletes well, start a new goroutine to delete components from here
	go Components.DeleteAllComponents(name)
	ctx.JSON(http.StatusOK, nil)
}
