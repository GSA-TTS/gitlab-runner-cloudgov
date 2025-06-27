package tutils

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"testing"

	cg "github.com/GSA-TTS/gitlab-runner-cloudgov/runner-manager/cfd/cloudgov"
)

const credPath = "./testdata/.cloudgov_creds"

func IntegrationSetup(t testing.TB) (client *cg.Client, org string, space string, app string) {
	var err error
	var user, pass string

	defer (func() {
		if err != nil {
			t.Fatal(fmt.Errorf("IntegrationSetup: %w", err))
		}
	})()

	cwd, err := os.Getwd()
	if err != nil {
		err = fmt.Errorf("getting cwd: %w", err)
		return
	}

	f, err := os.Open(path.Join(cwd, credPath))
	if err != nil {
		err = fmt.Errorf(
			"error opening testdata file = %v\n\033[1;33mDid you forget to create `%v`?\033[0m",
			err, credPath,
		)
		return
	}
	defer f.Close()

	vars := []*string{&user, &pass, &org, &space, &app}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()

		// Skipping comments
		if text[0] == '#' {
			continue
		}

		l := len(vars)
		if l != 0 {
			*vars[0] = text
			vars = vars[1:]
		} else {
			break
		}
	}

	if err = scanner.Err(); err != nil {
		err = fmt.Errorf("scanning testdata file: %w", err)
		return
	}

	if user == "" || pass == "" {
		err = fmt.Errorf("could not load variables from testdata")
		return
	}

	client, err = cg.New(&cg.CFClientAPI{}, &cg.Opts{
		Creds: &cg.Creds{Username: user, Password: pass},
	})
	if err != nil {
		err = fmt.Errorf("getting cloudgovClient: %w", err)
		return
	}

	return client, org, space, app
}

func CleanupApp(t testing.TB, c *cg.Client, guid string) {
	t.Helper()
	if err := c.AppDelete(guid); err != nil {
		t.Errorf("failed to delete app: %s", guid)
	}
}
