// Package cloudgov provides methods to interact CloudFoundry on cloud.gov.
package cloudgov

import "context"

type ClientAPI interface {
	connect(url string, creds *Creds) error

	appGet(id string) (*App, error)
	appPush(m *AppManifest) (*App, error)
	appDelete(id string) error
	appsList() (apps []*App, err error)

	sshCode() (string, error)
	mapRoute(ctx context.Context, app *App, domain string, space string, host string, path string, port int) error
	addNetworkPolicy(app *App, destGUID string, portRanges []string) error
}

type CredsGetter interface {
	getCreds() (*Creds, error)
}

type Opts struct {
	CredsGetter
	Creds *Creds

	APIRootURL string
}

type Client struct {
	ClientAPI
	*Opts
}

type CloudGovClientError struct {
	msg string
}

func (e CloudGovClientError) Error() string {
	return e.msg
}

// TODO: we should pull this out of VCAP_APPLICATION
const (
	apiRootURLDefault  = "https://api.fr-stage.cloud.gov"
	internalDomainGUID = "8a5d6a8c-cfc1-4fc4-afc9-aa563ff9df5e"
)

func New(i ClientAPI, o *Opts) (*Client, error) {
	if o == nil {
		o = &Opts{CredsGetter: EnvCredsGetter{}}
	}
	cg := &Client{ClientAPI: i, Opts: o}
	return cg.Connect()
}

func (c *Client) apiRootURL() string {
	if c.APIRootURL == "" {
		return apiRootURLDefault
	}
	return c.APIRootURL
}

func (c *Client) creds() (*Creds, error) {
	if c.Creds.isEmpty() {
		return c.getCreds()
	}
	return c.Creds, nil
}

func (c *Client) Connect() (*Client, error) {
	creds, err := c.creds()
	if err != nil {
		return nil, err
	}
	if err := c.connect(c.apiRootURL(), creds); err != nil {
		return nil, err
	}
	return c, nil
}

type App struct {
	Name      string
	GUID      string
	State     string
	SpaceGUID string
}

func (c *Client) AppGet(id string) (*App, error) {
	return c.appGet(id)
}

func (c *Client) AppDelete(id string) error {
	return c.appDelete(id)
}

func (c *Client) AppsList() ([]*App, error) {
	return c.appsList()
}

// TODO: this abstraction might belong in /cmd,
// unless it can be further generalized to all pushes
func (c *Client) Push(manifest *AppManifest) (*App, error) {
	containerID := manifest.Name

	if containerID == "" {
		return nil, CloudGovClientError{"Push: AppManifest.Name must be defined"}
	}

	if manifest.OrgName == "" || manifest.SpaceName == "" {
		return nil, CloudGovClientError{"Push: AppManifest must have Org and Space names"}
	}

	return c.appPush(manifest)
}

// TODO: use this in prepare or get rid of it
func (c *Client) ServicesPush(manifests []*AppManifest) ([]*App, error) {
	if len(manifests) < 1 {
		return nil, nil
	}

	apps := make([]*App, len(manifests))

	for i, s := range manifests {
		app, err := c.Push(s)
		if err != nil {
			return nil, err
		}
		apps[i] = app
	}

	return apps, nil
}

func (c *Client) SSHCode() (string, error) {
	return c.sshCode()
}

func (c *Client) MapServiceRoute(app *App) error {
	return c.mapRoute(context.Background(), app, internalDomainGUID, app.SpaceGUID, app.Name, "", 0)
}
