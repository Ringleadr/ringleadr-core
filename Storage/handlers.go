package Storage

import (
	"fmt"
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Datastore"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateStorage(ctx *gin.Context) {
	if ctx.Param("name") == "" {
		ctx.JSON(http.StatusInternalServerError, "must specify storage name")
		return
	}
	name := fmt.Sprintf("agogos-%s", ctx.Param("name"))

	err := Datastore.InsertStorage(&Datatypes.Storage{Name: name})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func DeleteStorage(ctx *gin.Context) {
	if ctx.Param("name") == "" {
		ctx.JSON(http.StatusInternalServerError, "must specify storage name")
		return
	}
	name := fmt.Sprintf("agogos-%s", ctx.Param("name"))

	err := Datastore.DeleteStorage(name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	//Trigger the deletion here as the watcher can't do it
	go Datastore.DeleteStorageInRuntime(name)

	ctx.JSON(http.StatusOK, nil)
}

func ListStorage(ctx *gin.Context) {
	storage, err := Datastore.GetAllStorage()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, storage)
}
