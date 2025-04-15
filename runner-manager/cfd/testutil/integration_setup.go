//go:build integration

package testutil

import (
	"bufio"
	"fmt"
	"os"
	"path"

	cg "github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cfd/cloudgov"
)

const credPath = "./testdata/.cloudgov_creds"

func IntegrationSetup() (client *cg.Client, err error) {
	var user, pass string

	defer (func() {
		if err != nil {
			err = fmt.Errorf("IntegrationSetup: %w", err)
		}
	})()

	cwd, err := os.Getwd()
	if err != nil {
		err = fmt.Errorf("getting cwd: %w", err)
	}

	fmt.Println(cwd)

	f, err := os.Open(path.Join(cwd, credPath))
	if err != nil {
		err = fmt.Errorf(
			"opening testdata file: %w\n\033[1;33mDid you forget to create `%v`?\033[0m",
			err, credPath,
		)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var i int
	var l [2]string
	for scanner.Scan() {
		text := scanner.Text()
		if text[0] == '#' {
			continue
		}
		l[i] = text
		if i++; i > 1 {
			user, pass = l[0], l[1]
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

	return client, nil
}
