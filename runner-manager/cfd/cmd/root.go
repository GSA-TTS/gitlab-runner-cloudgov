package cmd

import (
	"fmt"
	"os"

	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner-manager/cfd/cmd/drive"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(drive.DriveCmd)
}

var rootCmd = &cobra.Command{
	Use:   "cfd",
	Short: "CloudFoundry Driver",
	Long: `This is CloudFoundry Driver for the GitLab Custom executor.

The gitlab-runner service should run cfd with it's "drive" subcommands,
e.g., "cfd drive prepare".`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
