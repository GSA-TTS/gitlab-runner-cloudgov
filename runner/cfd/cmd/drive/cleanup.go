package drive

import (
	"log"

	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleans up resources, containers, etc. at job completion",
	Long: `The Cleanup stage is executed by "cleanup_exec".

This final stage is executed even if one of the previous stages failed.
The main goal for this stage is to clean up any of the environments that
might have been set up. For example, turning off VMs or deleting
containers.

The result of cleanup_exec does not affect job statuses. For example, a
job will be marked as successful even if the following occurs:

  Both prepare_exec and run_exec are successful.
  cleanup_exec fails.

The user can set cleanup_exec_timeout if they want to set some kind of
deadline of how long GitLab Runner should wait to clean up the
environment before terminating the process.

The STDOUT of this executable will be printed to GitLab Runner logs at a
DEBUG level. The STDERR will be printed to the logs at a WARN level.

Read more in GitLab's documentation:
https://docs.gitlab.com/runner/executors/custom.html#cleanup`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("cleaning up...")
	},
}
