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
	Run: run,
}

type prepStage commonStage

func run(cmd *cobra.Command, args []string) {
	// Move this stuff into a setup, add methods.
	s, err := newStage()
	if err != nil {
		panic(fmt.Errorf("error getting cgClient: %w", err))
	}

	s.prep.startServices()

	// if services, start services

	// if os.Getenv("")

	// create temp manifest

	// start container

	// install deps

	// allow access to services
}

func (s *prepStage) startServices() error {
	if len(s.config.Services) < 1 {
		return nil
	}

	for _, serv := range s.config.Services {
		containerID := fmt.Sprintf("%v-svc-%v", s.config.ContainerID, serv.Alias)

		// check for an old instance of the service, delete if found
		app, err := s.client.AppGet(containerID)
		if err != nil {
			return fmt.Errorf("error checking for existing service (%v): %w", containerID, err)
		}
		if app != nil {
			err = s.client.AppDelete(containerID)
		}
		if err != nil {
			return fmt.Errorf("error deleting existing service (%v): %w", containerID, err)
		}

		// args =
		// guid
		// worker_mem
		// docker image
		// health check type
		// no-route
		//
		// add job-vars & service-vars
		// (we add to $SVCMANIFEST right now, maybe a way around that?)
		//
		// add entrypoint & command
		//
		// add docker user/pass
		//
		// push
		//
		// map-route containerID apps.internal --hostname containerID
		//
		// export CI_SERVICE_$alias=$containerID.apps.internal
		//
	}

	return nil
}
