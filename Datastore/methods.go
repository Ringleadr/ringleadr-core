package Datastore

import (
	"github.com/GodlikePenguin/agogos-datatypes"
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
	err := storageCollection.Find(bson.M{"name": name}).One(network)
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		return nil, err
	}
	return network, err
}
