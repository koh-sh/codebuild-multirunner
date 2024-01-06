package cmd

import (
	"fmt"
	"log"

	cb "github.com/koh-sh/codebuild-multirunner/internal/codebuild"
	"github.com/spf13/cobra"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dump config for running CodeBuild projects",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := cb.DumpConfig(configfile)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(conf)
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
}
