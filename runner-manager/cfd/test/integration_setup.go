//go:build integration

package cfd_test

import (
	"bufio"
	"fmt"
	"os"

	cg "github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cfd/cloudgov"
)

func IntegrationSetup() (client *cg.Client, err error) {
	defer (func() {
		if err != nil {
			err = fmt.Errorf("IntegrationSetup error: %w", err)
		}
	})()

	var user, pass string

	path := "./testdata/.cloudgov_creds"
	f, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf(
			"opening testdata file: %w\n\033[1;33mDid you forget to create `%v`?\033[0m",
			err, path,
		)
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
