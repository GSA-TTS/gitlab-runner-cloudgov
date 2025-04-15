//go:build integration

package drive

import (
	"fmt"
	"os"
	"testing"

	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cfd/cloudgov"
	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cfd/testutil"
)

var cgClient *cloudgov.Client

func TestMain(m *testing.M) {
	var err error

	cgClient, err = testutil.IntegrationSetup()
	if err != nil {

		fmt.Printf("cloudgov_integration_test: %v\n", err)
		os.Exit(1)
	}

	m.Run()
}

func Test_RunSSH(t *testing.T) {
	var err error
	defer (func() {
		if err != nil {
			t.Fatal(err)
		}
	})()

	stage, err := newStage(cgClient)
	if err != nil {
		return
	}

	apps, err := cgClient.AppsList()
	if err != nil {
		return
	}

	fmt.Print(apps)

	err = stage.RunSSH(apps[0].GUID, "echo $VCAP_APPLICATION")
	if err != nil {
		return
	}
}
