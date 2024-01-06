package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// options

var (
	id         string
	nowait     bool
	pollsec    int
	configfile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "codebuild-multirunner",
	Short: "This is a simple CLI tool to \"Start build with overrides\" multiple AWS CodeBuild Projects at once.",
	Long: `This is a simple CLI tool to "Start build with overrides" multiple AWS CodeBuild Projects at once.

This command will read YAML based config file and run multiple CodeBuild projects with oneliner.
`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configfile, "config", "./.codebuild-multirunner.yaml", "file path for config file.")
}

// set version from goreleaser variables
func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)
}
