package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/google/go-cmp/cmp"
)

var syntaxError *json.SyntaxError

func getVcapJson(u string, p string) string {
	return fmt.Sprintf(`{"cloud-gov-service-account":[{"credentials":{"username":"%s","password":"%s"}}]}`, u, p)
}

func TestGetCfCredentials(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    *CfCredentials
		wantErr interface{}
	}{
		{
			"fails with no JSON",
			nil, nil, &syntaxError,
		},
		{
			"fails with malformed JSON",
			map[string]string{
				"VCAP_SERVICES": `{"foo": [{"bar": false}}`,
			},
			nil, &syntaxError,
		},
		{
			"fails with incorrectly defined VCAP envvar",
			map[string]string{
				"VCAP_SURGICES": getVcapJson("aa", "bb"),
			},
			nil, &syntaxError,
		},
		{
			"pulls credentials from JSON",
			map[string]string{
				"VCAP_SERVICES": getVcapJson("aa", "bb"),
			},
			&CfCredentials{"aa", "bb"}, nil,
		},
		{
			"pulls credentials from JSON when only user available",
			map[string]string{
				"CF_USERNAME":   "Klaus",
				"VCAP_SERVICES": getVcapJson("aa", "bb"),
			},
			&CfCredentials{"aa", "bb"}, nil,
		},
		{
			"pulls credentials from JSON when only pass available",
			map[string]string{
				"CF_PASSWORD":   "tulip-cat-cupcake",
				"VCAP_SERVICES": getVcapJson("aa", "bb"),
			},
			&CfCredentials{"aa", "bb"}, nil,
		},
		{
			"pulls credentials from specifically defined envvars if available",
			map[string]string{
				"CF_USERNAME":   "Klaus",
				"CF_PASSWORD":   "tulip-cat-cupcake",
				"VCAP_SERVICES": getVcapJson("aa", "bb"),
			},
			&CfCredentials{"Klaus", "tulip-cat-cupcake"}, nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			have, err := GetCfCredentials()

			if (err == nil) != (tt.wantErr == nil) {
				t.Errorf("GetCfClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && !errors.As(err, tt.wantErr) {
				t.Errorf("GetCredentials() bad error type: have %T, want %T", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, have); diff != "" {
				t.Error(diff)
			}
		})
	}
}

// func TestGetCfClient(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		want    *client.Client
// 		wantErr bool
// 	}{{}}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := GetCfClient()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetCfClient() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("GetCfClient() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_main(t *testing.T) {
	t.Skip("Test_main is just for experiments right now")

	tests := []struct{ name string }{{name: "run main"}}

	for _, tt := range tests {
		fmt.Println(tt.name)
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}
