package cmd

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dump config for running CodeBuild projects",
	Run: func(cmd *cobra.Command, args []string) {
		bc := readConfigFile(configfile)
		fmt.Println(dumpConfig(bc))
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
}

// dump read config with environment variables inserted
func dumpConfig(bc BuildConfig) string {
	d, err := yaml.Marshal(&bc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return string(d)
}
