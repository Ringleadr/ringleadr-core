package Datastore

import (
	"github.com/GodlikePenguin/agogos-datatypes"
	"github.com/GodlikePenguin/agogos-host/Containers"
	"github.com/globalsign/mgo/bson"
)

//CRUD actions

func InsertApp(application *Datatypes.Application) error {
	err := applicationCollection.Insert(application)
	if err != nil {
		return err
	}
	return nil
}

func UpdateApp(application *Datatypes.Application) error {
	err := applicationCollection.Update(bson.M{"name": application.Name}, application)
	if err != nil {
		return err
	}
	return nil
}

func GetApp(name string) (*Datatypes.Application, error) {
	app := &Datatypes.Application{}
	err := applicationCollection.Find(bson.M{"name": name}).One(app)
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		return &Datatypes.Application{}, err
	}
	return app, nil
}

func GetAllApps() ([]Datatypes.Application, error) {
	var apps []Datatypes.Application
	err := applicationCollection.Find(bson.M{}).All(&apps)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func DeleteApp(name string) error {
	//TODO check if app exists before deleting (otherwise this returns a blank error)
	err := applicationCollection.Remove(bson.M{"name": name})
	if err != nil {
		return err
	}
	return nil
}

func InsertStorage(storage *Datatypes.Storage) error {
	err := storageCollection.Insert(storage)
	if err != nil {
		return err
	}
	return nil
}

func DeleteStorage(name string) error {
	//TODO returns empty error when it can't delete the required item
	err := storageCollection.Remove(bson.M{"name": name})
	if err != nil {
		return err
	}
	return nil
}

func GetAllStorage() ([]Datatypes.Storage, error) {
	var storage []Datatypes.Storage
	err := storageCollection.Find(bson.M{}).All(&storage)
	if err != nil {
		return nil, err
	}
	return storage, nil
}

func GetStorage(name string) (*Datatypes.Storage, error) {
	storage := &Datatypes.Storage{}
	err := storageCollection.Find(bson.M{"name": name}).One(storage)
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		return nil, err
	}
	return storage, err
}

func InsertNetwork(network *Datatypes.Network) error {
	err := networkCollection.Insert(network)
	if err != nil {
		return err
	}
	return nil
}

func DeleteNetwork(name string) error {
	//TODO returns empty error when it can't delete the required item
	err := networkCollection.Remove(bson.M{"name": name})
	if err != nil {
		return err
	}
	return nil
}

func GetAllNetworks() ([]Datatypes.Network, error) {
	var networks []Datatypes.Network
	err := networkCollection.Find(bson.M{}).All(&networks)
	if err != nil {
		return nil, err
	}
	return networks, nil
}

func GetNetwork(name string) (*Datatypes.Network, error) {
	network := &Datatypes.Network{}
	err := networkCollection.Find(bson.M{"name": name}).One(network)
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		return nil, err
	}
	return network, err
}

func InsertNode(node *Datatypes.Node) error {
	err := nodesCollection.Insert(node)
	if err != nil {
		return err
	}
	return nil
}

func DeleteNode(name string) error {
	//TODO returns empty error when it can't delete the required item
	err := nodesCollection.Remove(bson.M{"name": name})
	if err != nil {
		return err
	}
	return nil
}

func GetAllNodes() ([]Datatypes.Node, error) {
	var nodes []Datatypes.Node
	err := nodesCollection.Find(bson.M{}).All(&nodes)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func GetNode(name string) (*Datatypes.Node, error) {
	node := &Datatypes.Node{}
	err := nodesCollection.Find(bson.M{"name": name}).One(node)
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		return nil, err
	}
	return node, err
}

func UpdateNode(node *Datatypes.Node) error {
	err := nodesCollection.Update(bson.M{"name": node.Name}, node)
	if err != nil {
		return err
	}
	return nil
}

// FIELDS FOR COMPONENT STATS

type ComponentEntry struct {
	AppName       string      `json:"app_name" bson:"app_name"`
	ComponentName string      `json:"component_name" bson:"component_name"`
	CpuUsage      []*CpuUsage `json:"cpu_usage" bson:"cpu_usage"`
}

type CpuUsage struct {
	TotalPercent   float64 `json:"total_percent" bson:"total_percent"`
	AveragePercent float64 `json:"average_percent" bson:"average_percent"`
	TimeStamp      int64   `json:"time_stamp" bson:"time_stamp"`
}

func UpdateOrInsertComponent(appName string, compName string, stats []Containers.Stats, timestamp int64) error {
	if len(stats) == 0 {
		return nil
	}
	var cpuTotal float64
	for _, stat := range stats {
		cpuTotal += stat.CpuUsage
	}
	usage := CpuUsage{TotalPercent: cpuTotal, AveragePercent: cpuTotal / float64(len(stats)), TimeStamp: timestamp}
	_, err := componentCollection.Upsert(bson.M{
		"app_name":       appName,
		"component_name": compName,
	}, bson.M{
		"$push": bson.M{
			"cpu_usage": usage,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func GetComponent(appName string, compName string) (*ComponentEntry, error) {
	comp := &ComponentEntry{}
	err := componentCollection.Find(bson.M{
		"app_name":       appName,
		"component_name": compName,
	}).One(comp)
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		return nil, err
	}
	return comp, err
}

func DeleteComponentsFor(appName string) error {
	err := componentCollection.Remove(bson.M{"app_name": appName})
	if err != nil {
		return err
	}
	return nil
}

type Overview struct {
	Applications int `json:"applications"`
	Networks     int `json:"networks"`
	Storage      int `json:"storage"`
	Nodes        int `json:"nodes"`
}

func GetOverview() (*Overview, error) {
	noApplications, err := applicationCollection.Count()
	if err != nil {
		return nil, err
	}
	noNetworks, err := networkCollection.Count()
	if err != nil {
		return nil, err
	}
	noStorage, err := storageCollection.Count()
	if err != nil {
		return nil, err
	}
	noNodes, err := nodesCollection.Count()
	if err != nil {
		return nil, err
	}
	return &Overview{
		Applications: noApplications,
		Networks:     noNetworks,
		Storage:      noStorage,
		Nodes:        noNodes,
	}, nil
}
