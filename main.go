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
		fmt.Println("error:", err)
		return nil, err
	}

	return &data.CloudGovServiceAccount[0].Credentials, nil
}

func main() {
	credentials, _ := GetCredentials()

	cfConfig, _ := config.New("https://api.fr.cloud.gov", config.UserPassword(
		credentials.Username,
		credentials.Password,
	))

	cf, _ := client.New(cfConfig)

	apps, _ := cf.Applications.ListAll(context.Background(), nil)
	for _, app := range apps {
		fmt.Printf("Application %s is %s\n", app.Name, app.State)
	}
}
