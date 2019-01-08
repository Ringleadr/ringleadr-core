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
		ctx.JSON(http.StatusInternalServerError, "must specify network name")
		return
	}
	name := fmt.Sprintf("agogos-%s", ctx.Param("name"))

	err := Datastore.InsertNetwork(&Datatypes.Network{Name: name})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func DeleteNetwork(ctx *gin.Context) {
	if ctx.Param("name") == "" {
		ctx.JSON(http.StatusInternalServerError, "must specify storage name")
		return
	}
	name := fmt.Sprintf("agogos-%s", ctx.Param("name"))

	err := Datastore.DeleteNetwork(name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	//Trigger the deletion here as the watcher can't do it
	go Datastore.DeleteNetworkInRuntime(name)

	ctx.JSON(http.StatusOK, nil)
}

func ListNetworks(ctx *gin.Context) {
	storage, err := Datastore.GetAllNetworks()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}
