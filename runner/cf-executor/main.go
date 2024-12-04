package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/joho/godotenv"
)

type CfCredentials struct {
	Username string
	Password string
}

type VcapData struct {
	CloudGovServiceAccount []struct {
		Credentials CfCredentials
	} `json:"cloud-gov-service-account"`
}

func GetCfCredentials() (*CfCredentials, error) {
	cfUser := os.Getenv("CF_USERNAME")
	cfPass := os.Getenv("CF_PASSWORD")

	if cfUser != "" && cfPass != "" {
		return &CfCredentials{cfUser, cfPass}, nil
	}

	var data VcapData
	vcapServices := os.Getenv("VCAP_SERVICES")

	if err := json.Unmarshal([]byte(vcapServices), &data); err != nil {
		return nil, fmt.Errorf("error unmarshaling VCAP_SERVICES: %w", err)
	}

	return &data.CloudGovServiceAccount[0].Credentials, nil
}

func GetCfClient(creds *CfCredentials) (_ *client.Client, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting cf client: %w", err)
		}
	}()

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

	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file")
	}

	creds, err := GetCfCredentials()
	if err != nil {
		return
	}

	cf, err := GetCfClient(creds)
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
