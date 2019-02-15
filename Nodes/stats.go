package Nodes

import (
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"net/http"
)

func Stats(ctx *gin.Context) {
	v, err := mem.VirtualMemory()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error getting mem: %s", err.Error())
		return
	}

	i, err := cpu.Info()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error getting cpu info: %s", err.Error())
		return
	}

	p, err := cpu.Percent(0, false)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error getting cpu usage: %s", err.Error())
		return
	}
	stats := &Datatypes.NodeStats{
		Cpus:       int(i[0].Cores),
		CpuPercent: p[0],
		TotalMem:   v.Total,
		UsedMem:    v.Used,
	}
	ctx.JSON(http.StatusOK, stats)
}
