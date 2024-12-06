package cg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type App struct {
	Id    string
	Name  string
	State string
}

type Creds struct {
	Username string
	Password string
}

func (cr *Creds) isEmpty() bool {
	return cr.Username == "" && cr.Password == ""
}

type vcapData struct {
	CGSrvAct []struct{ Creds Creds } `json:"cloud-gov-service-account"`
}

func getCreds() (*Creds, error) {
	creds := &Creds{}

	// Check if credentials are supplied through environment
	creds.Username = os.Getenv("CF_USERNAME")
	creds.Password = os.Getenv("CF_PASSWORD")
	if !creds.isEmpty() {
		return creds, nil
	}

	// Check for credentials in VCAP_SERVICES JSON
	var vcd vcapData
	vSrv := os.Getenv("VCAP_SERVICES")
	if err := json.Unmarshal([]byte(vSrv), &vcd); err != nil {
		return nil, fmt.Errorf("error unmarshaling VCAP_SERVICES: %w", err)
	}

	// If creds are still empty we fail
	creds = &vcd.CGSrvAct[0].Creds
	if creds.isEmpty() {
		return nil, errors.New("could not establish credentials")
	}

	return creds, nil
}

type Adapter interface {
	getApps() (apps []*App, err error)

	connect(url string, creds *Creds) error

	// Stuff
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
