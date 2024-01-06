package multirunner

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/aws/smithy-go/middleware"
)

//
// some types and functions for AWS SDK Mock. used only for testing
//

// mock api for StartBuild
type StartBuildMockAPI func(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error)

func (m StartBuildMockAPI) StartBuild(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error) {
	return m(ctx, params, optFns...)
}

func (m StartBuildMockAPI) BatchGetBuilds(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
	return nil, nil
}

func (m StartBuildMockAPI) RetryBuild(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error) {
	return nil, nil
}

// mock api for BatchGetBuilds
type BatchGetBuildsMockAPI func(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error)

func (m BatchGetBuildsMockAPI) StartBuild(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error) {
	return nil, nil
}

func (m BatchGetBuildsMockAPI) BatchGetBuilds(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
	return m(ctx, params, optFns...)
}

func (m BatchGetBuildsMockAPI) RetryBuild(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error) {
	return nil, nil
}

// mock api for BatchGetBuilds
type RetryBuildMockAPI func(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error)

func (m RetryBuildMockAPI) StartBuild(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error) {
	return nil, nil
}

func (m RetryBuildMockAPI) BatchGetBuilds(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
	return nil, nil
}

func (m RetryBuildMockAPI) RetryBuild(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error) {
	return m(ctx, params, optFns...)
}

// mock api for GetLogEvents
type GetLogEventsMockAPI func(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error)

func (m GetLogEventsMockAPI) GetLogEvents(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error) {
	return m(ctx, params, optFns...)
}

// return mock function for StartBuild
func ReturnStartBuildMockAPI(build *types.Build, err error) func(t *testing.T) CodeBuildAPI {
	return func(t *testing.T) CodeBuildAPI {
		return StartBuildMockAPI(func(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error) {
			t.Helper()
			// for error case
			if *params.ProjectName == "error" {
				return nil, fmt.Errorf("error")
			}
			return &codebuild.StartBuildOutput{
				Build:          build,
				ResultMetadata: middleware.Metadata{},
			}, nil
		})
	}
}

// return mock function for BatchgetBuilds
func ReturnBatchGetBuildsMockAPI(builds []types.Build) func(t *testing.T) CodeBuildAPI {
	return func(t *testing.T) CodeBuildAPI {
		return BatchGetBuildsMockAPI(func(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
			t.Helper()
			// for error case
			if params.Ids[0] == "error" {
				return nil, fmt.Errorf("error")
			}
			return &codebuild.BatchGetBuildsOutput{
				Builds:         builds,
				BuildsNotFound: []string{},
				ResultMetadata: middleware.Metadata{},
			}, nil
		})
	}
}

// return mock function for GetLogEvents
func ReturnGetLogEventsMockAPI(events []cwltypes.OutputLogEvent) func(t *testing.T) CWLGetLogEventsAPI {
	return func(t *testing.T) CWLGetLogEventsAPI {
		return GetLogEventsMockAPI(func(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error) {
			t.Helper()
			// for error case
			if *params.LogGroupName == "error" {
				return nil, fmt.Errorf("error")
			}
			return &cloudwatchlogs.GetLogEventsOutput{
				Events:            events,
				NextBackwardToken: nil,
				NextForwardToken:  nil,
			}, nil
		})
	}
}

// return mock function for BatchgetBuilds
func ReturnRetryBuildMockAPI(build types.Build) func(t *testing.T) CodeBuildAPI {
	return func(t *testing.T) CodeBuildAPI {
		return RetryBuildMockAPI(func(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error) {
			t.Helper()
			// for error case
			if *params.Id == "error" {
				return nil, fmt.Errorf("error")
			}
			return &codebuild.RetryBuildOutput{
				Build:          &build,
				ResultMetadata: middleware.Metadata{},
			}, nil
		})
	}
}
