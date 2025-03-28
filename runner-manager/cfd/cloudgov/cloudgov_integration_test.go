package cloudgov_test

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	cg "github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cfd/cloudgov"
	"github.com/google/go-cmp/cmp"
)

var (
	appGetWanted string
	cgClient     *cg.Client
)

func TestMain(m *testing.M) {
	var user, pass string
	var err error

	path := "./testdata/.cloudgov_creds"
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf(
			"Error opening testdata file = %v\n\033[1;33mDid you forget to create `%v`?\033[0m",
			err, path,
		)
		os.Exit(1)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var i int
	var l [3]string
	for scanner.Scan() {
		text := scanner.Text()
		if text[0] == '#' {
			continue
		}
		l[i] = text
		if i++; i > 2 {
			user, pass, appGetWanted = l[0], l[1], l[2]
			break
		}
	}

	if err = scanner.Err(); err != nil {
		fmt.Printf("Error scanning testdata file = %v", err)
		return
	}

	if user == "" || pass == "" {
		fmt.Printf("Could not load variables from testdata")
		return
	}

	cgClient, err = cg.New(&cg.CFClientAPI{}, &cg.Opts{
		Creds: &cg.Creds{Username: user, Password: pass},
	})
	if err != nil {
		fmt.Printf("Error getting cloudgovClient = %v", err)
		return
	}

	m.Run()
}

func Test_CFAdapter_AppGet(t *testing.T) {
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

func Test_ServicePush(t *testing.T) {
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
			want: &cg.App{Name: "c0f91804-4d3a-47df-be14-c9eb4fb59324", State: "STARTED"},
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
				t.Errorf("Client.ServicePush() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func Test_SSHCode(t *testing.T) {
	c := cgClient
	want := "hi"
	got, err := c.SSHCode()
	if err != nil {
		t.Errorf("got error = %v", err)
		return
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}
