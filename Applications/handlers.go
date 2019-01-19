package Applications

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Components"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Datastore"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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

	err = createApplication(app)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func createApplication(app *Datatypes.Application) error {
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
		return errors.New("an app already exists with that name")
	}

	err := Datastore.InsertApp(app)
	if err != nil {
		return err
	}
	return nil
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

func UpdateApplication(ctx *gin.Context) {
	//Update a specific application
	//Let's cheat and just delete the original app and create a new one
	//Get the application from the request body
	app := &Datatypes.Application{}
	err := ctx.BindJSON(app)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	code, err := deleteApplication(app.Name)
	if err != nil {
		ctx.JSON(code, err)
	}

	err = createApplication(app)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
	}

	ctx.JSON(http.StatusOK, nil)
}

func DeleteApplication(ctx *gin.Context) {
	name := ctx.Param("name")

	code, err := deleteApplication(name)
	if err != nil {
		ctx.JSON(code, err)
	}

	ctx.JSON(http.StatusOK, nil)
}

func deleteApplication(name string) (int, error) {
	app, err := getAppFromName(name)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = Datastore.DeleteApp(name)
	if err != nil {
		if err.Error() == "not found" {
			return http.StatusNotFound, errors.New(fmt.Sprintf("Application %s does not exist", name))
		}
		return http.StatusInternalServerError, err
	}
	//Changestreams don't handle deletes well, start a new goroutine to delete components from here
	//Delete the implicit application network
	go Components.DeleteAllComponents(name, app.Copies)
	go func() {
		retries := 5
		var err error
		for retries > 0 {
			if err = Containers.GetContainerRuntime().DeleteNetwork(app.Name); err == nil {
				return
			}
			time.Sleep(5 * time.Second)
			retries -= 1
		}
		Logger.ErrPrintf("Error deleting implicit network %s: %s", app.Name, err.Error())
	}()
	return http.StatusOK, nil
}
