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

	path := "./testdata/.cg_creds"
	f, err := os.Open(path)
	if err != nil {
		t.Errorf(
			"Error opening testdata file = %v\n\033[1;33mDid you forget to create `%v`?\033[0m",
			err, path,
		)
		return
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
		return
	}

	if u == "" || p == "" || want == "" {
		t.Error("Could not load variables from testdata")
		return
	}

	cgClient, err := cg.New(&cg.GoCFClientAdapter{}, &cg.Opts{
		Creds: &cg.Creds{Username: u, Password: p},
	})
	if err != nil {
		t.Errorf("Error getting cgClient = %v", err)
		return
	}

	apps, err := cgClient.GetApps()
	if err != nil {
		t.Errorf("Error running GetApps() = %v", err)
		return
	}

	got, err := json.Marshal(apps)
	if err != nil {
		t.Errorf("Error marshalling apps to json = %v", err)
		return
	}

	if diff := cmp.Diff(string(got), want); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
		return
	}
}
