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
	"os"
	"strings"
	"time"
)

//Create new application
func CreateApplication(ctx *gin.Context) {
	//Parse response into variable
	app := &Datatypes.Application{}
	err := ctx.BindJSON(app)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	err = createApplication(app)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func createApplication(app *Datatypes.Application) error {
	if app.Copies < 1 {
		app.Copies = 1
	}

	unnamedCounter := 0
	for _, comp := range app.Components {
		if comp.Name == "" {
			comp.Name = fmt.Sprintf("no-name-%d", unnamedCounter)
			unnamedCounter++
		}

		comp.Name = strings.Replace(comp.Name, "/", "_", -1)

		if comp.Replicas < 1 {
			comp.Replicas = 1
		}

		if comp.ScaleThreshold != 0 {
			if comp.ScaleMin < 1 {
				comp.ScaleMin = 1
			}

			if comp.ScaleMax < 1 {
				comp.ScaleMax = 10
			}
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
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, apps)
}

func GetApplication(ctx *gin.Context) {
	//Get a specific application
	appName := ctx.Param("name")
	if appName == "" {
		ctx.String(http.StatusInternalServerError, "must specify app name")
		return
	}

	app, err := getAppFromName(appName)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
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
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	code, err := deleteApplication(app.Name)
	if err != nil {
		ctx.String(code, err.Error())
		return
	}

	err = createApplication(app)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func DeleteApplication(ctx *gin.Context) {
	name := ctx.Param("name")

	code, err := deleteApplication(name)
	if err != nil {
		ctx.String(code, err.Error())
		return
	}

	err = Datastore.DeleteComponentsFor(name)
	if err != nil {
		//Not a crucial error so we won't return a non 200 code here, just log the error
		Logger.ErrPrintf("Could not remove components for %s: %s", name, err.Error())
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
	//We only need to issue these delete commands if the app is running on this node
	//If there is an error getting the hostname then just ignore it, as it should be cleaned up by the sync thread later anyway
	if host, err := os.Hostname(); err == nil && host == app.Node || "*" == app.Node {
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
	}
	return http.StatusOK, nil
}

func GetApplicationComponentInformation(ctx *gin.Context) {
	appName := ctx.Param("name")
	compName := ctx.Param("compName")
	comp, err := Datastore.GetComponent(appName, compName)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, comp)
}

func DeleteAllApps(ctx *gin.Context) {
	err := Datastore.DeleteAllApps()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error deleting all apps: %s", err.Error())
	}
	ctx.JSON(http.StatusOK, nil)
}

//Technically related to Components but easier to place it here to avoid import cycles
func DeleteAllComponentStats(ctx *gin.Context) {
	err := Datastore.DeleteCompStats()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error deleting all component stats: %s", err.Error())
		return
	}
	ctx.JSON(http.StatusOK, nil)
}
