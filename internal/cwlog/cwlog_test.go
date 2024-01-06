package cwlog

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/koh-sh/codebuild-multirunner/internal/cb"
)

// mock api for GetLogEvents
type GetLogEventsMockAPI func(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error)

func (m GetLogEventsMockAPI) GetLogEvents(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error) {
	return m(ctx, params, optFns...)
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

// return mock function for BatchgetBuilds
func ReturnBatchGetBuildsMockAPI(builds []types.Build) func(t *testing.T) cb.CodeBuildAPI {
	return func(t *testing.T) cb.CodeBuildAPI {
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

func Test_GetCloudWatchLogSetting(t *testing.T) {
	id := "project:12345678"
	group := "/aws/codebuild/project"
	stream := "12345678"
	type args struct {
		client func(t *testing.T) cb.CodeBuildAPI
		id     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name: "enabled",
			args: args{client: ReturnBatchGetBuildsMockAPI([]types.Build{
				{
					Logs: &types.LogsLocation{
						CloudWatchLogs: &types.CloudWatchLogsConfig{Status: "ENABLED"},
						GroupName:      &group,
						StreamName:     &stream,
					},
				},
			}), id: id},
			want:    group,
			want1:   stream,
			wantErr: false,
		},
		{
			name: "disabled",
			args: args{client: ReturnBatchGetBuildsMockAPI([]types.Build{
				{
					Logs: &types.LogsLocation{
						CloudWatchLogs: &types.CloudWatchLogsConfig{Status: "DISABLED"},
						GroupName:      &group,
						StreamName:     &stream,
					},
				},
			}), id: id},
			want:    "",
			want1:   "",
			wantErr: true,
		},
		{
			name:    "not found",
			args:    args{client: ReturnBatchGetBuildsMockAPI([]types.Build{}), id: id},
			want:    "",
			want1:   "",
			wantErr: true,
		},
		{
			name:    "api error",
			args:    args{client: ReturnBatchGetBuildsMockAPI([]types.Build{}), id: "error"},
			want:    "",
			want1:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetCloudWatchLogSetting(tt.args.client(t), tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCloudWatchLogSetting() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getCloudWatchLogSetting() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getCloudWatchLogSetting() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_GetCloudWatchLogEvents(t *testing.T) {
	lines := []string{"first line", "second line", "third line"}
	wantOutput := []cwltypes.OutputLogEvent{
		{
			IngestionTime: new(int64),
			Message:       &lines[0],
			Timestamp:     new(int64),
		},
		{
			IngestionTime: new(int64),
			Message:       &lines[1],
			Timestamp:     new(int64),
		},
		{
			IngestionTime: new(int64),
			Message:       &lines[2],
			Timestamp:     new(int64),
		},
	}
	type args struct {
		client func(t *testing.T) CWLGetLogEventsAPI
		group  string
		stream string
		token  string
	}
	tests := []struct {
		name    string
		args    args
		want    cloudwatchlogs.GetLogEventsOutput
		wantErr bool
	}{
		{
			name: "success",
			args: args{client: ReturnGetLogEventsMockAPI(wantOutput), group: "/aws/codebuild/project", stream: "12345678", token: ""},
			want: cloudwatchlogs.GetLogEventsOutput{
				Events:            wantOutput,
				NextBackwardToken: nil,
				NextForwardToken:  nil,
			},
			wantErr: false,
		},
		{
			name: "success with token",
			args: args{client: ReturnGetLogEventsMockAPI(wantOutput), group: "/aws/codebuild/project", stream: "12345678", token: "12345"},
			want: cloudwatchlogs.GetLogEventsOutput{
				Events:            wantOutput,
				NextBackwardToken: nil,
				NextForwardToken:  nil,
			},
			wantErr: false,
		},
		{
			name:    "empty group or stream",
			args:    args{client: ReturnGetLogEventsMockAPI([]cwltypes.OutputLogEvent{}), group: "", stream: "", token: ""},
			want:    cloudwatchlogs.GetLogEventsOutput{},
			wantErr: true,
		},
		{
			name:    "api error",
			args:    args{client: ReturnGetLogEventsMockAPI([]cwltypes.OutputLogEvent{}), group: "error", stream: "error", token: ""},
			want:    cloudwatchlogs.GetLogEventsOutput{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCloudWatchLogEvents(tt.args.client(t), tt.args.group, tt.args.stream, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCloudWatchLogEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCloudWatchLogEvents() = %v, want %v", got, tt.want)
			}
		})
	}
}
