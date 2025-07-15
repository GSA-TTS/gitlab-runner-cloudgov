package cloudgov

import (
	"context"
	"os"
	"strconv"
	"strings"

	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/policy_client"

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
	return &(App{Name: app.Name, GUID: app.GUID, State: app.State, SpaceGUID: app.Relationships.Space.Data.GUID})
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

func (cf *CFClientAPI) mapRoute(
	ctx context.Context,
	app *App,
	domain string, space string, host string, path string, port int,
) error {
	opts := resource.NewRouteCreateWithHost(domain, space, host, path, port)

	route, err := cf.conn().Routes.Create(ctx, opts)
	if err != nil {
		return err
	}

	_, err = cf.conn().Routes.InsertDestinations(
		ctx,
		route.GUID,
		[]*resource.RouteDestinationInsertOrReplace{{
			App: resource.RouteDestinationApp{GUID: &app.GUID},
		}},
	)
	return err
}

func parsePortRange(prange string) (start int, end int, err error) {
	ports := strings.Split(prange, "-")

	start, err = strconv.Atoi(ports[0])
	if err != nil {
		return
	}

	if len(ports) == 1 || ports[1] == "" {
		end = start
		return
	}
	end, err = strconv.Atoi(ports[1])
	return
}

func (cf *CFClientAPI) addNetworkPolicy(fromGUID string, toGUID string, portRanges []string) error {
	pclient := policy_client.NewExternal(
		lager.NewLogger("ExternalPolicyClient"),
		cf.conn().HTTPAuthClient(),
		cf.conn().ApiURL(""),
	)

	policies := make([]policy_client.Policy, len(portRanges))

	for i, prange := range portRanges {
		start, end, err := parsePortRange(prange)
		if err != nil {
			return err
		}

		policies[i] = policy_client.Policy{
			Source: policy_client.Source{ID: fromGUID},
			Destination: policy_client.Destination{
				ID:       toGUID,
				Ports:    policy_client.Ports{Start: start, End: end},
				Protocol: "tcp",
			},
		}
	}

	return pclient.AddPolicies("", policies)
}
