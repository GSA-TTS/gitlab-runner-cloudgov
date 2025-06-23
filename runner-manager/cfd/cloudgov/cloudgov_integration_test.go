//go:build integration

package cloudgov_test

import (
	"errors"
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
