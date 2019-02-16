package Nodes

import (
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/GodlikePenguin/agogos-host/Datastore"
	"github.com/GodlikePenguin/agogos-host/Logger"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"os"
	"time"
)

func StartStatsReporter() {
	go func() {
		for {
			hostname, err := os.Hostname()
			if err != nil {
				Logger.ErrPrintf("Error getting hostname in reporter thread: %s", err.Error())
			}
			time.Sleep(10 * time.Second)
			timestamp := time.Now().Unix()
			stats, err := getStats()
			if err != nil {
				Logger.ErrPrintf("Error getting stats in reporter thread: %s", err.Error())
				continue
			}
			stats.Timestamp = timestamp
			err = Datastore.UpdateOrInsertNodeStats(hostname, stats)
			if err != nil {
				Logger.ErrPrintf("Error updating stats in DB: %s", err.Error())
			}
		}
	}()
}

func getStats() (*Datatypes.NodeStats, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, errors.New("Error getting memory stats: " + err.Error())
	}

	i, err := cpu.Info()
	if err != nil {
		return nil, errors.New("Error getting CPU stats: " + err.Error())
	}

	p, err := cpu.Percent(0, false)
	if err != nil {
		return nil, errors.New("Error getting CPU Load: " + err.Error())
	}

	numConts, err := Containers.GetContainerRuntime().NumberOfContainers()
	if err != nil {
		return nil, errors.New("Error getting number of Containers: " + err.Error())
	}

	stats := &Datatypes.NodeStats{
		Cpus:           int(i[0].Cores),
		CpuPercent:     p[0],
		TotalMem:       v.Total,
		UsedMem:        v.Used,
		UsedMemPercent: v.UsedPercent,
		NumContainers:  numConts,
	}
	return stats, nil
}
