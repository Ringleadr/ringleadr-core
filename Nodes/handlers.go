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

func DeleteNode(ctx *gin.Context) {
	name := ctx.Param("name")
	node, err := Datastore.GetNode(name)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error searching for node: %s", err.Error())
		return
	}

	if node == nil {
		ctx.String(http.StatusNotFound, "No such node: %s", name)
		return
	}

	if node.Active == true {
		ctx.String(http.StatusBadRequest, "Cannot delete an active node")
		return
	}

	err = Datastore.DeleteNode(name)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error deleting node: %s", err.Error())
		return
	}

	err = Datastore.DeleteStatsFor(name)
	if err != nil {
		//Non critical error so we just log it and carry on
		Logger.ErrPrintf("Error removing stats for node %s: %s", name, err.Error())
	}
	ctx.JSON(http.StatusOK, nil)
}

func Stats(ctx *gin.Context) {
	name := ctx.Param("name")
	stats, err := Datastore.GetNodeStats(name)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Could not read stats for %s: %s", name, err.Error())
		return
	}
	if stats == nil {
		ctx.String(http.StatusNotFound, "No such node: %s", name)
		return
	}
	ctx.JSON(http.StatusOK, stats)
}
