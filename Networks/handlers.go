package Networks

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Datastore"
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

	err := Datastore.DeleteNetwork(name)
	if err != nil {
		if err.Error() == "not found" {
			ctx.String(http.StatusNotFound, "No such network %s", ctx.Param("name"))
			return
		}
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//Trigger the deletion here as the watcher can't do it
	go Datastore.DeleteNetworkInRuntime(name)

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
