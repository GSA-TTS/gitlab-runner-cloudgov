//go:build integration

package cloudgov_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	cg "github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cfd/cloudgov"
	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cfd/testutil"
	"github.com/google/go-cmp/cmp"
)

var cgClient *cg.Client

func TestMain(m *testing.M) {
	var err error

	cgClient, err = testutil.IntegrationSetup()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	m.Run()
}

func Test_CFAdapter_AppGet(t *testing.T) {
	appGetWanted := `[{"Name":"cgd-int-app","GUID":"0cfb2765-da96-4f0f-ad6f-f70cfa9400c2","State":"STARTED"},{"Name":"Some cool app","GUID":"1245231e-91cd-47da-ac5e-e22a1f624f9b","State":"STARTED"}]`
	apps, err := cgClient.AppsList()
	if err != nil {
		t.Errorf("Error running AppsList() = %v", err)
		return
	}

	got, err := json.Marshal(apps)
	if err != nil {
		t.Errorf("Error marshalling apps to json = %v", err)
		return
	}

	if diff := cmp.Diff(string(got), appGetWanted); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
		return
	}
}

func Test_Push(t *testing.T) {
	tests := map[string]struct {
		want     *cg.App
		wantErr  error
		manifest *cg.AppManifest
	}{
		"Fails with bad org & space": {
			wantErr:  errors.New("could not find org bad: expected exactly 1 result, but got less or more than 1"),
			manifest: &cg.AppManifest{Name: "Fail", OrgName: "bad", SpaceName: "bad"},
		},
		"Passes with real org and space": {
			want: &cg.App{Name: "Some cool app", GUID: "1245231e-91cd-47da-ac5e-e22a1f624f9b", State: "STARTED"},
			manifest: &cg.AppManifest{
				OrgName:   "gsa-tts-devtools-prototyping",
				SpaceName: "cgd-int",
				Name:      "Some cool app",
				Docker: cg.AppManifestDocker{
					Image: "busybox",
				},
				Process: cg.AppManifestProcess{
					HealthCheckType: "process",
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := cgClient
			got, err := c.Push(tt.manifest)
			if err != nil && (tt.wantErr == nil || err.Error() != tt.wantErr.Error()) {
				t.Errorf("Client.Push() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func Test_SSHCode(t *testing.T) {
	got, err := cgClient.SSHCode()
	if err != nil {
		t.Errorf("got error = %v", err)
		return
	}

	re := regexp.MustCompile(`[\w-_]{32}`)
	if !re.MatchString(got) {
		t.Errorf("wanted string matching /%v/, got %v", re, got)
	}
}
