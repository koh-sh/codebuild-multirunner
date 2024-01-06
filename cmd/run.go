package cmd

import (
	"log"
	"os"
	"time"

	cb "github.com/koh-sh/codebuild-multirunner/internal/codebuild"
	mr "github.com/koh-sh/codebuild-multirunner/internal/multirunner"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run CodeBuild projects based on YAML",
	Run: func(cmd *cobra.Command, args []string) {
		bc, err := mr.ReadConfigFile(configfile)
		if err != nil {
			log.Fatal(err)
		}
		client, err := cb.NewCodeBuildAPI()
		if err != nil {
			log.Fatal(err)
		}
		ids := []string{}
		hasfailedbuild := false
		for _, v := range bc.Builds {
			startbuildinput, err := mr.ConvertBuildConfigToStartBuildInput(v)
			if err != nil {
				log.Fatal(err)
			}
			id, err := cb.RunCodeBuild(client, startbuildinput)
			if err != nil {
				log.Println(err)
				hasfailedbuild = true
			} else {
				ids = append(ids, id)
			}
		}
		// early return if --no-wait option set
		if nowait {
			return
		}
		for i := 0; ; i++ {
			// break if all builds end
			if len(ids) == 0 {
				break
			}
			time.Sleep(time.Duration(pollsec) * time.Second)
			failed := false
			ids, failed, err = cb.BuildStatusCheck(client, ids)
			if err != nil {
				log.Fatal(err)
			}
			if failed {
				hasfailedbuild = true
			}
		}
		if hasfailedbuild {
			os.Exit(2)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&nowait, "no-wait", false, "specify if you don't need to follow builds status")
	runCmd.Flags().IntVar(&pollsec, "polling-span", 60, "polling span in second for builds status check")
}
