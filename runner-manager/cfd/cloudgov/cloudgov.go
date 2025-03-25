package cloudgov

// Stuff we'll need to implement, for ref
//
// mapRoute()
//
// addNetworkPolicy()
// removeNetworkPolicy()
type ClientAPI interface {
	connect(url string, creds *Creds) error

	appGet(id string) (*App, error)
	appPush(m *AppManifest) (*App, error)
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

type CloudGovClientError struct {
	msg string
}

func (e CloudGovClientError) Error() string {
	return e.msg
}

// TODO: we should pull this out of VCAP_APPLICATION
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
	Name  string
	GUID  string
	State string
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
func (c *Client) ServicePush(manifest *AppManifest) (*App, error) {
	containerID := manifest.Name

	if containerID == "" {
		return nil, CloudGovClientError{"ServicePush: AppManifest.Name must be defined"}
	}

	if manifest.OrgName == "" || manifest.SpaceName == "" {
		return nil, CloudGovClientError{"ServicePush: AppManifest must have Org and Space names"}
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
		app, err := c.ServicePush(s)
		if err != nil {
			return nil, err
		}
		apps[i] = app
	}

	return apps, nil
}
