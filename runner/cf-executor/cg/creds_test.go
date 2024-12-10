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
		want    *Creds
		wantErr interface{}
		env     map[string]string
		name    string
	}{
		{
			name:    "fails with no JSON",
			wantErr: &syntaxError,
		},
		{
			name:    "fails with malformed JSON",
			wantErr: &syntaxError,
			env: map[string]string{
				"VCAP_SERVICES": `{"foo": [{"bar": false}}`,
			},
		},
		{
			name:    "fails with incorrectly defined VCAP envvar",
			wantErr: &syntaxError,
			env: map[string]string{
				"VCAP_SURGICES": getVcapJson("aa", "bb"),
			},
		},
		{
			name: "pulls credentials from JSON",
			want: &Creds{Username: "aa", Password: "bb"},
			env: map[string]string{
				"VCAP_SERVICES": getVcapJson("aa", "bb"),
			},
		},
		{
			name: "pulls credentials from JSON when only user available",
			want: &Creds{Username: "aa", Password: "bb"},
			env: map[string]string{
				"CF_USERNAME":   "Klaus",
				"VCAP_SERVICES": getVcapJson("aa", "bb"),
			},
		},
		{
			name: "pulls credentials from JSON when only pass available",
			want: &Creds{Username: "aa", Password: "bb"},
			env: map[string]string{
				"CF_PASSWORD":   "tulip-cat-cupcake",
				"VCAP_SERVICES": getVcapJson("aa", "bb"),
			},
		},
		{
			name: "pulls credentials from specifically defined envvars if available",
			want: &Creds{Username: "Klaus", Password: "tulip-cat-cupcake"},
			env: map[string]string{
				"CF_USERNAME":   "Klaus",
				"CF_PASSWORD":   "tulip-cat-cupcake",
				"VCAP_SERVICES": getVcapJson("aa", "bb"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			have, err := EnvCredsGetter{}.getCreds()

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
