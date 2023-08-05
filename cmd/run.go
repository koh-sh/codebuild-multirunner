package cmd

import (
	"context"
	"log"
	"os"
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
			ids, failed = buildStatusCheck(client, ids)
			if failed {
				hasfailedbuild = true
			}
			// CodeBuild Timeout is 8h
			if pollsec*i > 8*60*60 {
				log.Fatal("Wait Timeout")
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

// run CodeBuild Projects and return build id
func runCodeBuild(client CodeBuildAPI, input codebuild.StartBuildInput) (string, error) {
	result, err := client.StartBuild(context.TODO(), &input)
	if err != nil {
		return "", err
	}
	id := *result.Build.Id
	log.Printf("%s [STARTED]\n", id)
	return id, err
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
func buildStatusCheck(client CodeBuildAPI, ids []string) ([]string, bool) {
	inprogressids := []string{}
	hasfailedbuild := false
	input := codebuild.BatchGetBuildsInput{Ids: ids}
	result, err := client.BatchGetBuilds(context.TODO(), &input)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range result.Builds {
		log.Printf("%s [%s]\n", *v.Id, coloredString(string(v.BuildStatus)))
		if v.BuildStatus == "IN_PROGRESS" {
			inprogressids = append(inprogressids, *v.Id)
		} else if v.BuildStatus != "SUCCEEDED" {
			hasfailedbuild = true
		}
	}
	return inprogressids, hasfailedbuild
}

// return colored string for each CodeBuild statuses
func coloredString(status string) string {
	if status == "SUCCEEDED" {
		return color.GreenString(status)
	} else if status == "IN_PROGRESS" {
		return color.BlueString(status)
	} else {
		return color.RedString(status)
	}
}
