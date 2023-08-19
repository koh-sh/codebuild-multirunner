package cmd

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/jinzhu/copier"
	root "github.com/koh-sh/codebuild-multirunner/cmd"
	"github.com/koh-sh/codebuild-multirunner/common"
	"github.com/spf13/cobra"
)

// options
var nowait bool
var pollsec int

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run CodeBuild projects based on YAML",
	Run: func(cmd *cobra.Command, args []string) {
		bc := common.ReadConfigFile(root.Configfile)
		client, err := common.NewCodeBuildAPI()
		if err != nil {
			log.Fatal(err)
		}
		ids := []string{}
		hasfailedbuild := false
		for _, v := range bc.Builds {
			startbuildinput := convertBuildConfigToStartBuildInput(v)
			id, err := runCodeBuild(client, startbuildinput)
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
			ids, failed = common.BuildStatusCheck(client, ids)
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
	root.RootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&nowait, "no-wait", false, "specify if you don't need to follow builds status")
	runCmd.Flags().IntVar(&pollsec, "polling-span", 60, "polling span in second for builds status check")

}

// run CodeBuild Projects and return build id
func runCodeBuild(client common.CodeBuildAPI, input codebuild.StartBuildInput) (string, error) {
	result, err := client.StartBuild(context.TODO(), &input)
	if err != nil {
		return "", err
	}
	id := *result.Build.Id
	log.Printf("%s [STARTED]\n", id)
	return id, err
}

// copy configration read from yaml to codebuild.StartBuildInput
func convertBuildConfigToStartBuildInput(build common.Build) codebuild.StartBuildInput {
	startbuildinput := codebuild.StartBuildInput{}
	copier.CopyWithOption(&startbuildinput, build, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	return startbuildinput
}
