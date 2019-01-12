package Applications

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Components"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Datastore"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

//Create new application
func CreateApplication(ctx *gin.Context) {
	//Parse response into variable
	app := &Datatypes.Application{}
	err := ctx.BindJSON(app)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if app.Copies < 1 {
		app.Copies = 1
	}

	for _, comp := range app.Components {
		if comp.Replicas < 1 {
			comp.Replicas = 1
		}
	}

	appExists, _ := getAppFromName(app.Name)
	if appExists != nil {
		ctx.JSON(http.StatusInternalServerError, "an app already exists with that name")
		return
	}

	err = Datastore.InsertApp(app)
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

func GetApplication(ctx *gin.Context) {
	//Get a specific application
	appName := ctx.Param("name")
	if appName == "" {
		ctx.JSON(http.StatusInternalServerError, "must specify app name")
		return
	}

	app, err := getAppFromName(appName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, app)
}

func getAppFromName(appName string) (*Datatypes.Application, error) {
	app, err := Datastore.GetApp(appName)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func UpdateApplication() {
	//Update a specific application
}

func DeleteApplication(ctx *gin.Context) {
	name := ctx.Param("name")

	app, err := getAppFromName(name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	err = Datastore.DeleteApp(name)
	if err != nil {
		if err.Error() == "not found" {
			ctx.JSON(http.StatusNotFound, fmt.Sprintf("Application %s does not exist", name))
			return
		}
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}
	//Changestreams don't handle deletes well, start a new goroutine to delete components from here
	//Delete the implicit application network
	go Components.DeleteAllComponents(name, app.Copies)
	go func() {
		retries := 5
		for retries > 0 {
			if err := Containers.GetContainerRuntime().DeleteNetwork(app.Name); err == nil {
				return
			}
			time.Sleep(5 * time.Second)
			retries -= 1
		}
	}()
	ctx.JSON(http.StatusOK, nil)
}
