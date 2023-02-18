package Applications

import (
	"errors"
	"fmt"
	"github.com/Ringleadr/ringleadr-core/internal/Datastore"
)

func RescheduleAppsOnNode(name string) error {
	apps, err := Datastore.GetAllApps()
	if err != nil {
		return err
	}
	for _, app := range apps {
		if app.Node == name {
			app.Node = ""
			err := Datastore.UpdateApp(&app)
			if err != nil {
				return errors.New(fmt.Sprintf("Could not reschedule app %s: %s", app.Name, err.Error()))
			}
		}
	}
	return nil
}
