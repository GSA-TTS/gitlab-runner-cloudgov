package drive

import (
	"log"

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
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("preparing...")
	},
}
