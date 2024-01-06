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
		bc, err := cb.ReadConfigFile(configfile)
		if err != nil {
			log.Fatal(err)
		}
		conf, err := cb.DumpConfig(bc)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(conf)
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
}
