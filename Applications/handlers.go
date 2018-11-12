package Applications

import (
	"github.com/GodlikePenguin/agogos-datatypes"
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
	}

	ctx.JSON(http.StatusOK, nil)
}

func GetApplications(ctx *gin.Context) {
	apps, err := Datastore.GetAllApps()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
	}

	ctx.JSON(http.StatusOK, apps)
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
