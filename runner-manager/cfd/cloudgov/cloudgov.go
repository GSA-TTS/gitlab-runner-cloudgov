package cloudgov

import "fmt"

// Stuff we'll need to implement, for ref
//
// mapRoute()
//
// addNetworkPolicy()
// removeNetworkPolicy()
//
// appGet()
// appCmd()
// appPush()
// appDelete()
type ClientAPI interface {
	connect(url string, creds *Creds) error

	appGet(id string) (*App, error)
	appPush(m AppManifest) (*App, error)
	appDelete(id string) error
	appsList() (apps []*App, err error)
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

const apiRootURLDefault = "https://api.fr.cloud.gov"

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
	Name  string // i.e., container ID
	State string
}

func (c *Client) AppGet(id string) (*App, error) {
	return c.appGet(id)
}

func (c *Client) AppDelete(id string) error {
	return c.appDelete(id)
}

// func (c *Client) JobPush(img *drive.Image, vars []*drive.CIVar)

type AppManifest struct {
	Name    string // i.e., container ID
	Env     map[string]string
	NoRoute bool
	Docker  *AppManifestDocker
	Process *AppManifestProcess
}
type AppManifestDocker struct {
	Image    string
	Username string
	Password string
}
type AppManifestProcess struct {
	Command         string // Entrypoint + Cmd
	DiskQuota       string
	Memory          string
	HealthCheckType string
}

func NewAppManifest(id string, memory string, disk string) *AppManifest {
	return &AppManifest{
		Name:    id,
		NoRoute: true,
		Docker:  &AppManifestDocker{},
		Process: &AppManifestProcess{
			Memory:          memory,
			DiskQuota:       disk,
			HealthCheckType: "process",
		},
	}
}

func (c *Client) ServicePush(manifest *AppManifest) (*App, error) {
	containerID := manifest.Name

	// check for an old instance of the service, delete if found
	app, err := c.AppGet(containerID)
	if err != nil {
		return nil, fmt.Errorf("error checking for existing service (%v): %w", containerID, err)
	}
	if app != nil {
		err = c.AppDelete(containerID)
	}
	if err != nil {
		return nil, fmt.Errorf("error deleting existing service (%v): %w", containerID, err)
	}

	return nil, nil
}

// func (c *Client) ServicesPush(svcs []*drive.Service) ([]*App, error) {
// 	if len(svcs) < 1 {
// 		return nil, nil
// 	}
//
// 	apps := make([]*App, len(svcs))
//
// 	for i, s := range svcs {
// 		app, err := c.ServicePush(s)
// 		if err != nil {
// 			return nil, err
// 		}
// 		apps[i] = app
// 	}
//
// 	return nil, nil
// }

func (c *Client) AppsList() ([]*App, error) {
	return c.appsList()
}
