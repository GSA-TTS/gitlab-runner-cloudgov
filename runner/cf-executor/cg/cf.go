package cg

import (
	"context"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
)

type GoCFClientAdapter struct {
	_con *client.Client
}

func (cf *GoCFClientAdapter) connect(url string, creds *Creds) error {
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

func (cf *GoCFClientAdapter) conn() *client.Client {
	if cf._con != nil {
		return cf._con
	}
	panic("not connected")
}

func (cf *GoCFClientAdapter) getApps() ([]*App, error) {
	apps, err := cf.conn().Applications.ListAll(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	Apps := make([]*App, len(apps))
	for idx, app := range apps {
		Apps[idx] = &(App{app.GUID, app.Name, app.State})
	}

	return Apps, nil
}
