package cloudgov

import (
	"context"
	"os"

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

func toCFManifest(am *AppManifest) *operation.AppManifest {
	return &operation.AppManifest{
		Name:    am.Name,
		Env:     am.Env,
		NoRoute: true,
		Docker: &operation.AppManifestDocker{
			Image:    am.Docker.Image,
			Username: am.Docker.Username,
		},
		AppManifestProcess: operation.AppManifestProcess{
			Command:         am.Process.Command,
			Memory:          am.Process.Memory,
			DiskQuota:       am.Process.DiskQuota,
			HealthCheckType: operation.AppHealthCheckType(am.Process.HealthCheckType),
		},
	}
}

// TODO: #95 - we'll want to change how docker creds get passed
func (cf *CFClientAPI) appPush(m *AppManifest) (*App, error) {
	// Initializes some state for the CF lib w/ connected client and org/space
	op := operation.NewAppPushOperation(cf._con, m.OrgName, m.SpaceName)

	// Translate manifest
	cfManifest := toCFManifest(m)

	// op.Push is a go-ified cli cmd, currently only takes env pass, see #95
	os.Setenv("CF_DOCKER_PASSWORD", m.Docker.Password)

	app, err := op.Push(context.Background(), cfManifest, nil)
	return castApp(app), err
}

func castApp(app *resource.App) *App {
	if app == nil || app.GUID == "" {
		return nil
	}
	return &(App{Name: app.Name, GUID: app.GUID, State: app.State})
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

func (cf *CFClientAPI) sshCode() (string, error) {
	ctx := context.Background()
	return cf.conn().SSHCode(ctx)
}
