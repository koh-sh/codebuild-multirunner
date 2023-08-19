package dump

import (
	"fmt"
	"log"

	root "github.com/koh-sh/codebuild-multirunner/cmd"
	"github.com/koh-sh/codebuild-multirunner/common"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dump config for running CodeBuild projects",
	Run: func(cmd *cobra.Command, args []string) {
		bc := common.ReadConfigFile(root.Configfile)
		fmt.Println(dumpConfig(bc))
	},
}

func init() {
	root.RootCmd.AddCommand(dumpCmd)
}

// dump read config with environment variables inserted
func dumpConfig(bc common.BuildConfig) string {
	d, err := yaml.Marshal(&bc)
	if err != nil {
		log.Fatal(err)
	}
	return string(d)
}
