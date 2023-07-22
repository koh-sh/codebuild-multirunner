package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run CodeBuild projects based on config",
	Run: func(cmd *cobra.Command, args []string) {
		bc := readConfigFile(configfile)
		runCodeBuild(bc)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

}

func runCodeBuild(bc BuildConfig) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	client := codebuild.NewFromConfig(cfg)
	for i := 0; i < len(bc.Builds); i++ {
		startbuildinput := convertBuildConfigToStartBuildInput(bc.Builds[i])
		result, err := client.StartBuild(context.TODO(), &startbuildinput)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		fmt.Println(result)
    // TODO: test
	}
}

// copy configration read from yaml to codebuild.StartBuildInput
func convertBuildConfigToStartBuildInput(build Build) codebuild.StartBuildInput {
	startbuildinput := codebuild.StartBuildInput{}
	copier.CopyWithOption(&startbuildinput, build, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	return startbuildinput
}
