package drive

import (
	"log"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure various jobs settings before they run",
	Long: `The Config stage is executed by "config_exec".

Sometimes you might want to set some settings during execution time. For
example, setting a build directory depending on the project ID.

For a detailed list of settings that can be configured, read more at:
https://docs.gitlab.com/runner/executors/custom.html#config.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("configuring...")
	},
}
