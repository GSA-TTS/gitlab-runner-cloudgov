package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
)

var vcap = []byte(`{"cloud-gov-service-account":[{"label":"cloud-gov-service-account","provider":null,"plan":"space-deployer","name":"zjr-gl-runner-svc","tags":["gitlab-service-account"],"instance_guid":"aaaa","instance_name":"zjr-gl-runner-svc","binding_guid":"c55db941-d3eb-482c-af75-62a37503bc88","binding_name":null,"credentials":{"password":"REDACTED","username":"REDACTED"}}]}`)

type VCapData struct {
	CloudGovServiceAccount []struct {
		Label       string
		Credentials struct {
			Username string
			Password string
		}
	} `json:"cloud-gov-service-account"`
}

func main() {
	var vcapData VCapData

	err := json.Unmarshal(vcap, &vcapData)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	credentials := vcapData.CloudGovServiceAccount[0].Credentials

	cfConfig, _ := config.New("https://api.fr.cloud.gov", config.UserPassword(
		credentials.Username,
		credentials.Password,
	))
	cf, _ := client.New(cfConfig)

	fmt.Printf("%+v", vcapData)

	apps, _ := cf.Applications.ListAll(context.Background(), nil)
	for _, app := range apps {
		fmt.Printf("Application %s is %s\n", app.Name, app.State)
	}
}
