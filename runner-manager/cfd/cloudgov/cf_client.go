package cloudgov

import (
	"context"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/cloudfoundry/go-cfclient/v3/operation"
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

// TODO: this isn't a great name
func toExtManifest(am *AppManifest) *operation.AppManifest {
	return operation.NewManifest().Applications[0]
}

// TODO: several…
// - process org and space options
// - translate manifest from internal to external form
// - possibly track a proper context
// - get docker pass into env (CF_DOCKER_PASSWORD) if used
func (cf *CFClientAPI) appPush(m *AppManifest) (*App, error) {
	op := operation.NewAppPushOperation(cf._con, "org", "space")
	app, err := op.Push(context.Background(), toExtManifest(m), nil)
	return castApp(app), err
}

func castApp(app *resource.App) *App {
	return &(App{Name: app.GUID, State: app.State})
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
