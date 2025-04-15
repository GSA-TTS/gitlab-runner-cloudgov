package drive

import (
	"fmt"

	"github.com/spf13/cobra"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare for jobs by starting containers, services, etc.",
	Long: `The Prepare stage is executed by "prepare_exec".

At this point, GitLab Runner knows everything about the job (where and
how it’s going to run). The only thing left is for the environment to be
set up so the job can run. Prepare will execute the steps necessary to
create that environment.

This is responsible for setting up the environment (for example,
creating the virtual machine or container, services or anything else).
After this is done, we expect that the environment is ready to run the
job.

This stage is executed only once in a job execution.

The STDOUT and STDERR returned from this executable will print to the
job log.

Read more in GitLab's documentation:
https://docs.gitlab.com/runner/executors/custom.html#prepare`,
	RunE: run,
}

type prepStage commonStage

func run(cmd *cobra.Command, args []string) error {
	s, err := newStage(nil)
	if err != nil {
		return fmt.Errorf("error initializing prepare stage: %w", err)
	}

	err = s.prep.exec()
	if err != nil {
		return fmt.Errorf("error executing prepare stage: %w", err)
	}

	return nil
}

func (s *prepStage) exec() (err error) {
	// Looping service manifests to run `cf push`
	err = s.startServices()
	if err != nil {
		return err
	}

	// Pushing the main job config pulled from get_job_config.go
	app, err := s.client.Push(s.config.Manifest)
	if err != nil {
		return err
	}

	s.RunSSH(app.GUID,
		`source /etc/profile && (command -v git && command -v git-lfs && command -v curl) || \
        (command -v apk && https_proxy=$http_proxy apk add git git-lfs curl) || \
        (command -v apt-get && echo "Acquire::http::Proxy \"$http_proxy\";" > /etc/apt/apt.conf.d/proxy.conf && apt-get update && apt-get install -y git git-lfs curl) || \
        (command -v dnf && dnf -y install git git-lfs curl) || \
        (echo "[cf-driver] Required packages missing and install attempt failed" && exit 1)'`,
	)

	// TODO:
	// allow access to services

	return err
}

// TODO: refactor to include a service manifests slice and
// use client.ServicesPush or get rid of it
func (s *prepStage) startServices() error {
	if len(s.config.Services) < 1 {
		return nil
	}

	for _, serv := range s.config.Services {
		s.client.Push(serv.Manifest)
		// map-route containerID apps.internal --hostname containerID
		// export CI_SERVICE_$alias=$containerID.apps.internal
	}

	return nil
}
