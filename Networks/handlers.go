package Networks

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Datastore"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/GodlikePenguin/agogos-host/Utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateNetwork(ctx *gin.Context) {
	if ctx.Param("name") == "" {
		ctx.String(http.StatusInternalServerError, "must specify network name")
		return
	}
	name := fmt.Sprintf("agogos-%s", ctx.Param("name"))

	err := Datastore.InsertNetwork(&Datatypes.Network{Name: name})
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func DeleteNetwork(ctx *gin.Context) {
	if ctx.Param("name") == "" {
		ctx.String(http.StatusInternalServerError, "must specify storage name")
		return
	}
	name := fmt.Sprintf("agogos-%s", ctx.Param("name"))

	apps, err := Datastore.GetAllApps()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error checking if network can be deleted: %s", err.Error())
		return
	}

	for _, app := range apps {
		if Utils.StringArrayContains(app.Networks, ctx.Param("name")) {
			ctx.String(http.StatusBadRequest, "Cannot delete network %s as it is in use by application %s."+
				" Please delete this application between deleting the network.", ctx.Param("name"), app.Name)
			return
		}
	}

	err = Datastore.DeleteNetwork(name)
	if err != nil {
		if err.Error() == "not found" {
			ctx.String(http.StatusNotFound, "No such network %s", ctx.Param("name"))
			return
		}
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//Trigger the deletion here as the watcher can't do it
	go func() {
		err := Datastore.DeleteNetworkInRuntime(name)
		if err != nil {
			Logger.ErrPrintf("Error deleting network in runtime: %s", err.Error())
		}
	}()

	ctx.JSON(http.StatusOK, nil)
}

func ListNetworks(ctx *gin.Context) {
	storage, err := Datastore.GetAllNetworks()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

func DeleteAllNetworks(ctx *gin.Context) {
	err := Datastore.DeleteAllNetworks()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error deleting networks: %s", err.Error())
		return
	}
	ctx.JSON(http.StatusOK, nil)
}
