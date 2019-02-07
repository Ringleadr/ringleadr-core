package Nodes

import (
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Datastore"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func ListNodes(ctx *gin.Context) {
	nodes, err := Datastore.GetAllNodes()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, nodes)
}

func Register(ctx *gin.Context) {
	req := &struct {
		Name string `json:"name"`
	}{}
	address := ctx.Request.RemoteAddr[:strings.LastIndex(ctx.Request.RemoteAddr, ":")]
	err := ctx.ShouldBindJSON(req)
	if err != nil {
		Logger.ErrPrintln("Error binding register request: ", err)
		ctx.String(http.StatusInternalServerError, "Could not get name from register request")
		return
	}
	if node, err := Datastore.GetNode(req.Name); err == nil && node == nil {
		err = Datastore.InsertNode(&Datatypes.Node{Name: req.Name, Address: address})
		if err != nil {
			Logger.ErrPrintln("Error registering new node: ", err)
			ctx.String(http.StatusInternalServerError, "Error inserting new node: "+err.Error())
			return
		}
	} else if err != nil {
		Logger.ErrPrintf("Error checking for existing node entry: %s", err.Error())
		ctx.String(http.StatusInternalServerError, "Error checking for existing node entry: %s"+err.Error())
		return
	}
	ctx.JSON(http.StatusOK, nil)
}
