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

func TestGetCredentials(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    *Credentials
		wantErr interface{}
	}{
		{"fails with no JSON", "", nil, &syntaxError},
		{"fails with malformed JSON", `{"foo": [{"bar": false}}`, nil, &syntaxError},
		{"pulls credentials from JSON", getVcapJson("aa", "bb"), &Credentials{"aa", "bb"}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("VCAP_SERVICES", tt.json)

			have, err := GetCredentials()

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

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "scoobie"},
	}
	for _, tt := range tests {
		fmt.Println(tt.name)
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}
