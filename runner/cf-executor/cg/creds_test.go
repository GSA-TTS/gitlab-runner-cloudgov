package cg

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var syntaxError *json.SyntaxError

func getVcapJson(u string, p string) string {
	return fmt.Sprintf(`{"cloud-gov-service-account":[{"credentials":{"username":"%s","password":"%s"}}]}`, u, p)
}

func Test_getCreds(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    *Creds
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
			&Creds{Username: "aa", Password: "bb"}, nil,
		},
		{
			"pulls credentials from JSON when only user available",
			map[string]string{
				"CF_USERNAME":   "Klaus",
				"VCAP_SERVICES": getVcapJson("aa", "bb"),
			},
			&Creds{Username: "aa", Password: "bb"}, nil,
		},
		{
			"pulls credentials from JSON when only pass available",
			map[string]string{
				"CF_PASSWORD":   "tulip-cat-cupcake",
				"VCAP_SERVICES": getVcapJson("aa", "bb"),
			},
			&Creds{Username: "aa", Password: "bb"}, nil,
		},
		{
			"pulls credentials from specifically defined envvars if available",
			map[string]string{
				"CF_USERNAME":   "Klaus",
				"CF_PASSWORD":   "tulip-cat-cupcake",
				"VCAP_SERVICES": getVcapJson("aa", "bb"),
			},
			&Creds{Username: "Klaus", Password: "tulip-cat-cupcake"}, nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			have, err := getCreds()

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
