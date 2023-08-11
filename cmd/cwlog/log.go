package cwlog

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	root "github.com/koh-sh/codebuild-multirunner/cmd"
	"github.com/spf13/cobra"
)

var id string

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Print CodeBuild log for a single build with a provided id.",
	Long: `Print CodeBuild log for a single build with a provided id.

Only CloudWatch Logs is supported.
S3 Log is not supported`,

	Run: func(cmd *cobra.Command, args []string) {
		cbclient, err := root.NewCodeBuildAPI()
		if err != nil {
			log.Fatal(err)
		}
		group, stream, err := getCloudWatchLogSetting(cbclient, id)
		if err != nil {
			log.Fatal(err)
		}
		cwlclient, err := root.NewCloudWatchLogsAPI()
		if err != nil {
			log.Fatal(err)
		}
		res, err := getCloudWatchLogEvents(cwlclient, group, stream)
		if err != nil {
			log.Fatal(err)
		}
		for _, event := range res.Events {
			fmt.Print(*event.Message)
		}
	},
}

func init() {
	root.RootCmd.AddCommand(logCmd)
	logCmd.Flags().StringVar(&id, "id", "", "CodeBuild build id for getting log")
	logCmd.MarkFlagRequired("id")
}

// get CloudWatch Log settings from a build and return logGroupName, logStreamName and error
func getCloudWatchLogSetting(client root.CodeBuildAPI, id string) (string, string, error) {
	input := codebuild.BatchGetBuildsInput{Ids: []string{id}}
	result, err := client.BatchGetBuilds(context.TODO(), &input)
	if err != nil {
		return "", "", err
	}
	if len(result.Builds) == 0 {
		return "", "", fmt.Errorf("%v is not found", id)
	}
	build := result.Builds[0].Logs
	if build.CloudWatchLogs.Status == "DISABLED" {
		return "", "", fmt.Errorf("CloudWatch Logs for %v is Disabled", id)
	}
	return *build.GroupName, *build.StreamName, nil
}

// get CloudWatchLog events and return GetLogEventsOutput
func getCloudWatchLogEvents(client root.CWLGetLogEventsAPI, group string, stream string) (cloudwatchlogs.GetLogEventsOutput, error) {
	if group == "" || stream == "" {
		return cloudwatchlogs.GetLogEventsOutput{}, errors.New("you must supply a logGroupName and logStreamName")
	}
	input := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  &group,
		LogStreamName: &stream,
	}
	result, err := client.GetLogEvents(context.TODO(), input)
	if err != nil {
		return cloudwatchlogs.GetLogEventsOutput{}, err
	}
	return *result, nil
}
