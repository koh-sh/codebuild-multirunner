package cmd

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/fatih/color"
	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
)

// options
var nowait bool
var pollsec int

// interface for AWS API mock
type CodeBuildAPI interface {
	BatchGetBuilds(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error)
	StartBuild(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error)
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run CodeBuild projects based on YAML",
	Run: func(cmd *cobra.Command, args []string) {
		bc := readConfigFile(configfile)
		client, err := NewAPI()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		ids := runCodeBuild(client, bc)
		if !nowait {
			for i := 0; ; i++ {
				time.Sleep(time.Duration(pollsec) * time.Second)
				ids = buildStatusCheck(client, ids)
				// break if all builds end
				if len(ids) == 0 {
					break
				}
				// CodeBuild Timeout is 8h
				if pollsec*i > 8*60*60 {
					log.Fatal("Wait Timeout")
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&nowait, "no-wait", false, "specify if you don't need to follow builds status")
	runCmd.Flags().IntVar(&pollsec, "polling-span", 60, "polling span in second for builds status check")

}

// run CodeBuild Projects and return build ids
func runCodeBuild(client CodeBuildAPI, bc BuildConfig) []string {
	ids := []string{}
	for i := 0; i < len(bc.Builds); i++ {
		startbuildinput := convertBuildConfigToStartBuildInput(bc.Builds[i])
		result, err := client.StartBuild(context.TODO(), &startbuildinput)
		if err != nil {
			log.Println(err)
			continue
		}
		id := *result.Build.Id
		ids = append(ids, id)
		log.Printf("%s [STARTED]\n", id)
	}
	return ids
}

// return api client
func NewAPI() (CodeBuildAPI, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	return codebuild.NewFromConfig(cfg), nil
}

// copy configration read from yaml to codebuild.StartBuildInput
func convertBuildConfigToStartBuildInput(build Build) codebuild.StartBuildInput {
	startbuildinput := codebuild.StartBuildInput{}
	copier.CopyWithOption(&startbuildinput, build, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	return startbuildinput
}

// check builds status and return ongoing build ids
func buildStatusCheck(client CodeBuildAPI, ids []string) []string {
	inprogressids := []string{}
	input := codebuild.BatchGetBuildsInput{Ids: ids}
	result, err := client.BatchGetBuilds(context.TODO(), &input)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(result.Builds); i++ {
		if result.Builds[i].BuildStatus == "SUCCEEDED" {
			log.Printf("%s [%s]\n", *result.Builds[i].Id, color.GreenString(string(result.Builds[i].BuildStatus)))
		} else if result.Builds[i].BuildStatus == "IN_PROGRESS" {
			log.Printf("%s [%s]\n", *result.Builds[i].Id, color.BlueString(string(result.Builds[i].BuildStatus)))
			inprogressids = append(inprogressids, *result.Builds[i].Id)
		} else {
			log.Printf("%s [%s]\n", *result.Builds[i].Id, color.RedString(string(result.Builds[i].BuildStatus)))
		}
	}
	return inprogressids
}
