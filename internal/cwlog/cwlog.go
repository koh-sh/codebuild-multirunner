package cwlog

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	mr "github.com/koh-sh/codebuild-multirunner/internal/multirunner"
)

// get CloudWatch Log settings from a build and return logGroupName, logStreamName and error
func GetCloudWatchLogSetting(client mr.CodeBuildAPI, id string) (string, string, error) {
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
func GetCloudWatchLogEvents(client mr.CWLGetLogEventsAPI, group string, stream string, token string) (cloudwatchlogs.GetLogEventsOutput, error) {
	startfromhead := true
	if group == "" || stream == "" {
		return cloudwatchlogs.GetLogEventsOutput{}, errors.New("you must supply a logGroupName and logStreamName")
	}
	input := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  &group,
		LogStreamName: &stream,
		StartFromHead: &startfromhead,
	}
	if token != "" {
		input.NextToken = &token
	}
	result, err := client.GetLogEvents(context.TODO(), input)
	if err != nil {
		return cloudwatchlogs.GetLogEventsOutput{}, err
	}
	return *result, nil
}
