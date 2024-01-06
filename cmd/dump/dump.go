package dump

import (
	"fmt"
	"log"

	root "github.com/koh-sh/codebuild-multirunner/cmd"
	mr "github.com/koh-sh/codebuild-multirunner/internal/multirunner"
	"github.com/spf13/cobra"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dump config for running CodeBuild projects",
	Run: func(cmd *cobra.Command, args []string) {
		bc, err := mr.ReadConfigFile(root.Configfile)
		if err != nil {
			log.Fatal(err)
		}
		conf, err := mr.DumpConfig(bc)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(conf)
	},
}

func init() {
	root.RootCmd.AddCommand(dumpCmd)
}
