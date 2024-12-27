package drive

import (
	"github.com/spf13/cobra"
)

func init() {
	cobra.EnableCommandSorting = false
	DriveCmd.AddCommand(configCmd, prepareCmd, runCmd, cleanupCmd)
}

var DriveCmd = &cobra.Command{
	Use:   "drive",
	Short: "Drive stages requested by gitlab-runner's executor",
	Long: `Drive holds subcommands to run each of executor's stages.

The Custom executor provides the stages for you to configure some
details of the job, prepare and clean up the environment and run the job
script within it. Each stage is responsible for specific things and has
different things to keep in mind.

Each stage executed by the Custom executor is executed at the time a
builtin GitLab Runner executor would execute them.

For each step that will be executed, specific environment variables are
exposed to the executable, which can be used to get information about the
specific Job that is running. All stages will have the following
environment variables available to them:

  - Standard CI/CD environment variables, including predefined variables.
  - All environment variables provided by the Custom executor Runner host
    system.
  - All services and their available settings. Exposed in JSON format as
    CUSTOM_ENV_CI_JOB_SERVICES.

Both CI/CD environment variables and predefined variables are prefixed
with CUSTOM_ENV_ to prevent conflicts with system environment variables.
For example, CI_BUILDS_DIR will be available as CUSTOM_ENV_CI_BUILDS_DIR.

The stages run in the following sequence:

  1. config_exec
  2. prepare_exec
  3. run_exec
  4. cleanup_exec

Read more in GitLab's documentation:
https://docs.gitlab.com/runner/executors/custom.html#stages
  `,
}
