//go:build integration

package cg_test

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"

	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cf-driver/cg"
	"github.com/google/go-cmp/cmp"
)

func Test_CFAdapter_GetApps(t *testing.T) {
	var u, p, want string
	var l int

	f, err := os.Open("./testdata/.cg_creds")
	if err != nil {
		t.Errorf("Error opening testdata file = %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

scanning:
	for scanner.Scan() {
		text := scanner.Text()

		if text[0] == '#' {
			continue
		}

		switch l {
		case 0:
			u = text
		case 1:
			p = text
		case 2:
			want = text
			break scanning
		}

		l++
	}

	if err := scanner.Err(); err != nil {
		t.Errorf("Error scanning testdata file = %v", err)
	}

	if u == "" || p == "" || want == "" {
		t.Error("Could not load variables from testdata")
	}

	cgClient, err := cg.New(&cg.GoCFClientAdapter{}, &cg.Opts{
		Creds: &cg.Creds{Username: u, Password: p},
	})
	if err != nil {
		t.Errorf("Error getting cgClient = %v", err)
	}

	apps, err := cgClient.GetApps()
	if err != nil {
		t.Errorf("Error running GetApps() = %v", err)
	}

	got, err := json.Marshal(apps)
	if err != nil {
		t.Errorf("Error marshalling apps to json = %v", err)
	}

	if diff := cmp.Diff(string(got), want); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}
