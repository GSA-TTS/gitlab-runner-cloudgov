//go:build integration

package drive

import (
	"os/exec"
	"testing"

	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner-manager/cfd/cloudgov"
	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner-manager/cfd/internal/tutils"
)

var (
	cgClient *cloudgov.Client
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

func Test_RunSSH(t *testing.T) {
	setup(t)

	stage, err := newStage(cgClient)
	if err != nil {
		t.Fatal(err)
	}

	apps, err := cgClient.AppsList()
	if err != nil {
		t.Fatal(err)
	}

	err = stage.RunSSH(apps[0].GUID, "echo $VCAP_APPLICATION")
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatal(string(exitErr.Stderr))
		} else {
			t.Fatal(err)
		}
	}
}
