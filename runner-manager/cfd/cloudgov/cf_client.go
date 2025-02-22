package cloudgov

import (
	"context"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
)

type CFClientAPI struct {
	_con *client.Client
}

func (cf *CFClientAPI) connect(url string, creds *Creds) error {
	cfg, err := config.New(url, config.UserPassword(creds.Username, creds.Password))
	if err != nil {
		return err
	}

	con, err := client.New(cfg)
	if err != nil {
		return err
	}

	cf._con = con
	return nil
}

func (cf *CFClientAPI) conn() *client.Client {
	if cf._con != nil {
		return cf._con
	}
	panic("go-cfclient adapter is not connected")
}

func castApp(app *resource.App) *App {
	return &(App{app.GUID, app.Name, app.State})
}

func castApps(apps []*resource.App) []*App {
	Apps := make([]*App, len(apps))
	for idx, app := range apps {
		Apps[idx] = castApp(app)
	}
	return Apps
}

func (cf *CFClientAPI) appGet(id string) (*App, error) {
	app, err := cf.conn().Applications.Get(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return castApp(app), nil
}

func (cf *CFClientAPI) appDelete(id string) error {
	_, err := cf.conn().Applications.Delete(context.Background(), id)
	return err
}

func (cf *CFClientAPI) appsList() ([]*App, error) {
	apps, err := cf.conn().Applications.ListAll(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return castApps(apps), nil
}
