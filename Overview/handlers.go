package Overview

import (
	"github.com/GodlikePenguin/agogos-host/Datastore"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetOverview(ctx *gin.Context) {
	overview, err := Datastore.GetOverview()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, overview)
}

func Purge(ctx *gin.Context) {
	err := Datastore.DeleteAllApps()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error deleting apps: %s", err.Error())
	}
	err = Datastore.DeleteAllNetworks()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error deleting apps: %s", err.Error())
	}
	err = Datastore.DeleteAllStorage()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error deleting apps: %s", err.Error())
	}
	ctx.JSON(http.StatusOK, nil)
}
