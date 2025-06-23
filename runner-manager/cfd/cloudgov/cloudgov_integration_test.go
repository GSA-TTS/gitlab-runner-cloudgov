//go:build integration

package cloudgov_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"testing"

	cg "github.com/GSA-TTS/gitlab-runner-cloudgov/runner-manager/cfd/cloudgov"
	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner-manager/cfd/internal/tutils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	cgClient *cg.Client
	app,
	org,
	space string
)

func setup(t testing.TB) {
	t.Helper()
	if cgClient != nil {
		return
	}
	cgClient, org, space, app = tutils.IntegrationSetup(t)
}

func getCmpOpts() cmp.Option {
	return cmpopts.IgnoreFields(cg.App{}, "GUID")
}

func Test_CFAdapter_AppGet(t *testing.T) {
	setup(t)

	want := []*cg.App{{
		Name:  app,
		State: "STARTED",
	}}

	got, err := cgClient.AppsList()
	if err != nil {
		t.Fatalf("Error running AppsList() = %v", err)
	}

	if diff := cmp.Diff(got, want, getCmpOpts()); diff != "" {
		t.Fatalf("mismatch (-got +want):\n%s", diff)
	}
}

func Test_Push(t *testing.T) {
	setup(t)

	tests := map[string]struct {
		want     *cg.App
		wantErr  error
		manifest *cg.AppManifest
	}{
		"Fails with bad org & space": {
			wantErr:  errors.New("could not find org bad: expected exactly 1 result, but got less or more than 1"),
			manifest: &cg.AppManifest{Name: "Fail", OrgName: "bad", SpaceName: "bad"},
		},
		"Passes with sandbox space": {
			want: &cg.App{Name: "Test_Push_App", State: "STARTED"},
			manifest: &cg.AppManifest{
				OrgName:   org,
				SpaceName: space,
				Name:      "Test_Push_App",
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

			if got != nil && got.GUID != "" {
				tutils.CleanupApp(t, c, got.GUID)
			}

			if err != nil && (tt.wantErr == nil || err.Error() != tt.wantErr.Error()) {
				t.Errorf("Client.Push() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(got, tt.want, getCmpOpts()); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func Test_SSHCode(t *testing.T) {
	setup(t)
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

func cleanupRoute(t testing.TB, app *cg.App) error {
	t.Helper()

	delRouteCmd := exec.Command(
		"cf", "delete-route", "-f", "apps.internal",
		fmt.Sprintf("-n%s", app.Name),
	)

	out, err := delRouteCmd.CombinedOutput()
	if err != nil {
		t.Log(string(out))
		if exErr, ok := err.(*exec.ExitError); ok {
			t.Log(exErr.Error())
			t.Fatal(string(exErr.Stderr))
		} else {
			t.Fatal(err)
		}
	}

	return err
}

func TestClient_MapServiceRoute(t *testing.T) {
	setup(t)

	apps, err := cgClient.AppsList()
	if err != nil {
		t.Fatal(err)
	}
	app := apps[0]

	err = cgClient.MapServiceRoute(app)
	defer cleanupRoute(t, app)
	if err != nil {
		t.Fatal(err)
	}

	ckRouteCmd := exec.Command("cf", "curl", fmt.Sprintf("/v3/apps/%s/routes", app.GUID))
	out, err := ckRouteCmd.CombinedOutput()
	if err != nil {
		t.Log(out)
		t.Fatal(err)
	}

	var routeOut map[string][]map[string]string
	if err := json.Unmarshal(out, &routeOut); err != nil {
		t.Log("partial unmarshalling error expected…")
		t.Log(err)
	}

	wantUrl := fmt.Sprintf("%s.apps.internal", app.Name)

	for _, m := range routeOut["resources"] {
		if m["host"] == app.Name && m["url"] == wantUrl {
			return
		}
	}

	t.Logf("%#v", routeOut["resources"])
	t.Fatalf("could not find route with %s host and correct url", app.Name)
}
