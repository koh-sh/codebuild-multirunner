package cmd

import (
	"log"
	"os"

	"github.com/koh-sh/runcbs/internal/cb"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run CodeBuild projects based on YAML",
	Run: func(cmd *cobra.Command, args []string) {
		bc, err := cb.ReadConfigFile(configfile)
		if err != nil {
			log.Fatal(err)
		}
		client, err := cb.NewCodeBuildAPI()
		if err != nil {
			log.Fatal(err)
		}
		// run specified codebuild projects
		ids := []string{}
		runfailed := false
		for _, v := range bc.Builds {
			input, err := cb.ConvertBuildConfigToStartBuildInput(v)
			if err != nil {
				log.Fatal(err)
			}
			id, err := cb.RunCodeBuild(client, input)
			if err != nil {
				log.Println(err)
				runfailed = true
			} else {
				ids = append(ids, id)
			}
		}
		// early return if --no-wait option set
		if nowait {
			return
		}
		// check build status
		failed := false
		failed, err = cb.WaitAndCheckBuildStatus(client, ids, pollsec)
		if err != nil {
			log.Fatal(err)
		}
		if runfailed || failed {
			os.Exit(2)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&nowait, "no-wait", false, "specify if you don't need to follow builds status")
	runCmd.Flags().IntVar(&pollsec, "polling-span", 60, "polling span in second for builds status check")
}
