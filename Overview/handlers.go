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
