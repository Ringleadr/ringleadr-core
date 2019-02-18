package Storage

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Datastore"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateStorage(ctx *gin.Context) {
	if ctx.Param("name") == "" {
		ctx.String(http.StatusInternalServerError, "must specify storage name")
		return
	}
	name := fmt.Sprintf("agogos-%s", ctx.Param("name"))

	err := Datastore.InsertStorage(&Datatypes.Storage{Name: name})
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func DeleteStorage(ctx *gin.Context) {
	if ctx.Param("name") == "" {
		ctx.String(http.StatusInternalServerError, "must specify storage name")
		return
	}
	name := fmt.Sprintf("agogos-%s", ctx.Param("name"))

	apps, err := Datastore.GetAllApps()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error checking if storage can be deleted: %s", err.Error())
		return
	}

	for _, app := range apps {
		for _, comp := range app.Components {
			for _, store := range comp.Storage {
				if store.Name == ctx.Param("name") {
					ctx.String(http.StatusBadRequest, "Cannot delete storage %s as it is in use by component %s"+
						" in application %s. Please delete this application between deleting the network.",
						ctx.Param("name"), comp.Name, app.Name)
					return
				}
			}
		}
	}

	err = Datastore.DeleteStorage(name)
	if err != nil {
		if err.Error() == "not found" {
			ctx.String(http.StatusNotFound, "No such storage %s", ctx.Param("name"))
			return
		}
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//Trigger the deletion here as the watcher can't do it
	go func() {
		err := Datastore.DeleteStorageInRuntime(name)
		if err != nil {
			Logger.ErrPrintf("Error deleting storage %s in runtime: %s", name, err)
		}

	}()

	ctx.JSON(http.StatusOK, nil)
}

func ListStorage(ctx *gin.Context) {
	storage, err := Datastore.GetAllStorage()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

func DeleteAllStorage(ctx *gin.Context) {
	err := Datastore.DeleteAllStorage()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error deleting storage: %s", err.Error())
		return
	}
	ctx.JSON(http.StatusOK, nil)
}
