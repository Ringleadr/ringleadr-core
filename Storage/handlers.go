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

	err := Datastore.DeleteStorage(name)
	if err != nil {
		if err.Error() == "not found" {
			ctx.String(http.StatusNotFound, "No such storage %s", ctx.Param("name"))
			return
		}
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//Trigger the deletion here as the watcher can't do it
	go Datastore.DeleteStorageInRuntime(name)

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
