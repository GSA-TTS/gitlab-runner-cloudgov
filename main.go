package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
)

type Credentials struct {
	Username string
	Password string
}

type VcapData struct {
	CloudGovServiceAccount []struct {
		Credentials Credentials
	} `json:"cloud-gov-service-account"`
}

func GetCredentials() (*Credentials, error) {
	var data VcapData
	vcapServices := os.Getenv("VCAP_SERVICES")

	err := json.Unmarshal([]byte(vcapServices), &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling VCAP_SERVICES: %w", err)
	}

	return &data.CloudGovServiceAccount[0].Credentials, nil
}

func GetCfClient() (_ *client.Client, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting cf client: %w", err)
		}
	}()

	creds, err := GetCredentials()
	if err != nil {
		return nil, err
	}

	apiRootUrl := "https://api.fr.cloud.gov"
	configOpts := config.UserPassword(creds.Username, creds.Password)
	cfConfig, err := config.New(apiRootUrl, configOpts)
	if err != nil {
		return nil, err
	}

	return client.New(cfConfig)
}

func main() {
	var err error

	defer func() {
		if err != nil {
			fmt.Println(fmt.Errorf("in main: %w", err))
		}

		if v := recover(); v != nil {
			fmt.Println(v)
		}
	}()

	cf, err := GetCfClient()
	if err != nil {
		return
	}

	apps, err := cf.Applications.ListAll(context.Background(), nil)
	if err != nil {
		return
	}

	for _, app := range apps {
		fmt.Printf("Application %s is %s\n", app.Name, app.State)
	}
}
