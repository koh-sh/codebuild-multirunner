package common

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

//
// types and functions shared within subcommands
//

// interface for AWS CodeBuild API
type CodeBuildAPI interface {
	BatchGetBuilds(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error)
	StartBuild(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error)
	RetryBuild(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error)
}

// interface for AWS CloudWatch Logs API
type CWLGetLogEventsAPI interface {
	GetLogEvents(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error)
}

// return CodeBuild api client
func NewCodeBuildAPI() (CodeBuildAPI, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	return codebuild.NewFromConfig(cfg), nil
}

// return CloudWatchLogs api client
func NewCloudWatchLogsAPI() (CWLGetLogEventsAPI, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	return cloudwatchlogs.NewFromConfig(cfg), nil
}

// read yaml config file for builds definition
func ReadConfigFile(filepath string) BuildConfig {
	bc := BuildConfig{}
	b, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	expanded := os.ExpandEnv(string(b))
	err = yaml.Unmarshal([]byte(expanded), &bc)
	if err != nil {
		log.Fatal(err)
	}
	return bc
}

// check builds status and return ongoing build ids
func BuildStatusCheck(client CodeBuildAPI, ids []string) ([]string, bool) {
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
	switch status {
	case "SUCCEEDED":
		return color.GreenString(status)
	case "IN_PROGRESS":
		return color.BlueString(status)
	default:
		return color.RedString(status)
	}
}
