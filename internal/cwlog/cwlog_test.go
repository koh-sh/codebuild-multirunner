package cwlog

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
)

type MockCodeBuildAPI struct {
	StartBuildMock     func(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error)
	BatchGetBuildsMock func(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error)
	RetryBuildMock     func(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error)
}

func (m *MockCodeBuildAPI) StartBuild(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error) {
	return m.StartBuildMock(ctx, params, optFns...)
}

func (m *MockCodeBuildAPI) BatchGetBuilds(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
	return m.BatchGetBuildsMock(ctx, params, optFns...)
}

func (m *MockCodeBuildAPI) RetryBuild(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error) {
	return m.RetryBuildMock(ctx, params, optFns...)
}

func NewMockCodeBuildAPI(startBuildMock func(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error),
	batchGetBuildsMock func(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error),
	retryBuildMock func(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error),
) *MockCodeBuildAPI {
	return &MockCodeBuildAPI{
		StartBuildMock:     startBuildMock,
		BatchGetBuildsMock: batchGetBuildsMock,
		RetryBuildMock:     retryBuildMock,
	}
}

type MockCWLGetLogEventsAPI struct {
	GetLogEventsMock func(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error)
}

func (m *MockCWLGetLogEventsAPI) GetLogEvents(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error) {
	return m.GetLogEventsMock(ctx, params, optFns...)
}

func NewMockCWLGetLogEventsAPI(getLogEventsMock func(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error)) *MockCWLGetLogEventsAPI {
	return &MockCWLGetLogEventsAPI{
		GetLogEventsMock: getLogEventsMock,
	}
}

func TestGetCloudWatchLogSetting(t *testing.T) {
	mockCodeBuildAPI := NewMockCodeBuildAPI(
		nil,
		func(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
			if params.Ids[0] == "error:12345678" {
				return nil, errors.New("batch get builds error")
			}
			var builds []types.Build
			switch params.Ids[0] {
			case "project:12345678":
				builds = []types.Build{
					{
						Logs: &types.LogsLocation{
							CloudWatchLogs: &types.CloudWatchLogsConfig{Status: "ENABLED"},
							GroupName:      aws.String("/aws/codebuild/project"),
							StreamName:     aws.String("12345678"),
						},
					},
				}
			case "project2:12345678":
				builds = []types.Build{
					{
						Logs: &types.LogsLocation{
							CloudWatchLogs: &types.CloudWatchLogsConfig{Status: "DISABLED"},
							GroupName:      aws.String("/aws/codebuild/project"),
							StreamName:     aws.String("12345678"),
						},
					},
				}
			default:
				builds = []types.Build{}
			}
			return &codebuild.BatchGetBuildsOutput{
				Builds: builds,
			}, nil
		},
		nil,
	)

	tests := []struct {
		name    string
		id      string
		want    string
		want1   string
		wantErr bool
	}{
		{
			name:    "CloudWatch Logs enabled",
			id:      "project:12345678",
			want:    "/aws/codebuild/project",
			want1:   "12345678",
			wantErr: false,
		},
		{
			name:    "CloudWatch Logs disabled",
			id:      "project2:12345678",
			want:    "",
			want1:   "",
			wantErr: true,
		},
		{
			name:    "Build not found",
			id:      "project3:12345678",
			want:    "",
			want1:   "",
			wantErr: true,
		},
		{
			name:    "API error",
			id:      "error:12345678",
			want:    "",
			want1:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetCloudWatchLogSetting(mockCodeBuildAPI, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCloudWatchLogSetting() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetCloudWatchLogSetting() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetCloudWatchLogSetting() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGetCloudWatchLogEvents(t *testing.T) {
	mockCWLGetLogEventsAPI := NewMockCWLGetLogEventsAPI(
		func(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error) {
			if *params.LogGroupName == "error" {
				return nil, errors.New("get log events error")
			}
			var events []cwltypes.OutputLogEvent
			switch *params.LogGroupName {
			case "/aws/codebuild/project":
				events = []cwltypes.OutputLogEvent{
					{
						Message: aws.String("first line"),
					},
					{
						Message: aws.String("second line"),
					},
					{
						Message: aws.String("third line"),
					},
				}
			default:
				events = []cwltypes.OutputLogEvent{}
			}
			return &cloudwatchlogs.GetLogEventsOutput{
				Events: events,
			}, nil
		},
	)

	tests := []struct {
		name    string
		group   string
		stream  string
		token   string
		want    cloudwatchlogs.GetLogEventsOutput
		wantErr bool
	}{
		{
			name:   "Get log events successfully",
			group:  "/aws/codebuild/project",
			stream: "12345678",
			token:  "",
			want: cloudwatchlogs.GetLogEventsOutput{
				Events: []cwltypes.OutputLogEvent{
					{
						Message: aws.String("first line"),
					},
					{
						Message: aws.String("second line"),
					},
					{
						Message: aws.String("third line"),
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "Get log events with token",
			group:  "/aws/codebuild/project",
			stream: "12345678",
			token:  "12345",
			want: cloudwatchlogs.GetLogEventsOutput{
				Events: []cwltypes.OutputLogEvent{
					{
						Message: aws.String("first line"),
					},
					{
						Message: aws.String("second line"),
					},
					{
						Message: aws.String("third line"),
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Empty group or stream",
			group:   "",
			stream:  "",
			token:   "",
			want:    cloudwatchlogs.GetLogEventsOutput{},
			wantErr: true,
		},
		{
			name:    "API error",
			group:   "error",
			stream:  "error",
			token:   "",
			want:    cloudwatchlogs.GetLogEventsOutput{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCloudWatchLogEvents(mockCWLGetLogEventsAPI, tt.group, tt.stream, tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCloudWatchLogEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCloudWatchLogEvents() = %v, want %v", got, tt.want)
			}
		})
	}
}
