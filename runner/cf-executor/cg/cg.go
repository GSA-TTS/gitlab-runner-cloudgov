package cg

type App struct {
	Id    string
	Name  string
	State string
}

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
type Adapter interface {
	getApps() (apps []*App, err error)
	connect(url string, creds *Creds) error
}

type CG struct {
	Adapter
	Opts
}

type Opts struct {
	Creds      Creds
	APIRootURL string
}

var apiRootURLDefault = "https://api.fr.cloud.gov"

func New(a Adapter, o *Opts) (*CG, error) {
	if o == nil {
		o = &Opts{}
	}
	cg := &CG{a, *o}
	return cg.Connect()
}

func (c *CG) apiRootURL() string {
	if c.APIRootURL == "" {
		return apiRootURLDefault
	}
	return c.APIRootURL
}

func (c *CG) creds() (*Creds, error) {
	if c.Creds.isEmpty() {
		return getCreds()
	}
	return &c.Creds, nil
}

func (c *CG) Connect() (*CG, error) {
	creds, err := c.creds()
	if err != nil {
		return nil, err
	}
	if err := c.connect(c.apiRootURL(), creds); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *CG) GetApps() ([]*App, error) {
	return c.getApps()
}
