package drive

import (
	"log"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Executes substages (e.g., cache download) and job steps",
	Long: `The Run stage is executed by "run_exec".

The STDOUT and STDERR returned from this executable will print to the job
log.

Unlike the other stages, the run_exec stage is executed multiple times,
since it’s split into sub stages listed below in sequential order:

  prepare_script
  get_sources
  restore_cache
  download_artifacts
  step_*
  build_script
  step_*
  after_script
  archive_cache OR archive_cache_on_failure
  upload_artifacts_on_success OR upload_artifacts_on_failure
  cleanup_file_variables

In GitLab Runner 14.0 and later, build_script will be replaced with
step_script. For more information, see this issue.

For each stage mentioned above, the run_exec executable will be executed
with (1) the path to the script that GitLab Runner creates for the Custom
executor to run, and (2) the name of the stage.

This executable should be responsible for executing the scripts that are
specified in the first argument. They contain all the scripts any GitLab
Runner executor would run normally to clone, download artifacts, run
user scripts and all the other steps described below. The scripts can be
of the following shells:

  Bash
  PowerShell Desktop
  PowerShell Core
  Batch (deprecated)

We generate the script using the shell configured by shell inside of
[[runners]]. If none is provided the defaults for the OS platform are used.

The table below is a detailed explanation of what each script does and
what the main goal of that script is.

Read more in GitLab's documentation:
https://docs.gitlab.com/runner/executors/custom.html#run`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("running...")
	},
}
