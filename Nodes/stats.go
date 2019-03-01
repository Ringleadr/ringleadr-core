package Nodes

import (
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Datastore"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"net/http"
	"os"
	"time"
)

func StartStatsReporter() {
	go func() {
		errorMap := make(map[string]interface{})
		counter := 0
		for {
			time.Sleep(10 * time.Second)
			counter++
			if counter == 5*6 {
				//If we've run for 5 minutes, then empty the error map so we can report errors again
				errorMap = make(map[string]interface{})
				counter = 0
			}
			hostname, err := os.Hostname()
			if err != nil {
				Logger.ErrPrintf("Error getting hostname in reporter thread: %s", err.Error())
				continue
			}
			timestamp := time.Now().Unix()
			stats, errs := getStats()
			if len(errs) > 0 {
				for _, err := range errs {
					if _, ok := errorMap[err.Error()]; !ok {
						Logger.ErrPrintf("Could not get all stats in reporter thread: %s", err.Error())
						//Doesn't matter what we assign it, we never use the value
						errorMap[err.Error()] = true
					} else {
						//Already seen this error, don't print it again yet.
					}
				}
			}
			stats.Timestamp = timestamp
			err = Datastore.UpdateOrInsertNodeStats(hostname, stats)
			if err != nil {
				Logger.ErrPrintf("Error updating stats in DB: %s", err.Error())
			}
		}
	}()
}

func getStats() (*Datatypes.NodeStats, []error) {
	var errorMessages []error
	var unavailable []string
	v, err := mem.VirtualMemory()
	if err != nil {
		errorMessages = append(errorMessages, errors.New("Error getting memory stats: "+err.Error()))
		v = &mem.VirtualMemoryStat{
			Total:       0,
			Used:        0,
			UsedPercent: 0,
		}
		unavailable = append(unavailable, "Virtual Memory")
	}

	i, err := cpu.Info()
	if err != nil {
		errorMessages = append(errorMessages, errors.New("Error getting CPU stats: "+err.Error()))
		i = []cpu.InfoStat{
			{
				Cores: 0,
			},
		}
		unavailable = append(unavailable, "CPU Cores")
	}

	p, err := cpu.Percent(0, false)
	if err != nil {
		errorMessages = append(errorMessages, errors.New("Error getting CPU Load: "+err.Error()))
		p = []float64{0}
		unavailable = append(unavailable, "CPU Load")
	}

	numConts, err := Containers.GetContainerRuntime().NumberOfContainers()
	if err != nil {
		errorMessages = append(errorMessages, errors.New("Error getting number of Containers: "+err.Error()))
		numConts = 0
		unavailable = append(unavailable, "Running Containers")
	}

	stats := &Datatypes.NodeStats{
		Cpus:           int(i[0].Cores),
		CpuPercent:     p[0],
		TotalMem:       v.Total,
		UsedMem:        v.Used,
		UsedMemPercent: v.UsedPercent,
		NumContainers:  numConts,
		Unavailable:    unavailable,
	}
	return stats, errorMessages
}

func DeleteAllStats(ctx *gin.Context) {
	err := Datastore.DeleteAllNodeStats()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Error deleting node stats: %s", err.Error())
	}
	ctx.JSON(http.StatusOK, nil)
}
