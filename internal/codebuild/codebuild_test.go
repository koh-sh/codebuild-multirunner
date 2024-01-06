package codebuild

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/aws/smithy-go/middleware"
)

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

func Test_BuildStatusCheck(t *testing.T) {
	id1 := "project:12345678"
	id2 := "project2:87654321"
	ids := []string{id1, id2}
	errids := []string{"error"}

	type args struct {
		client func(t *testing.T) CodeBuildAPI
		ids    []string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		want2   bool
		wantErr bool
	}{
		{
			name:    "all builds ended",
			args:    args{client: ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "SUCCEEDED", Id: &id2}}), ids: ids},
			want:    []string{},
			want2:   false,
			wantErr: false,
		},
		{
			name:    "one builds in progress",
			args:    args{client: ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "IN_PROGRESS", Id: &id2}}), ids: ids},
			want:    []string{id2},
			want2:   false,
			wantErr: false,
		},
		{
			name:    "one of builds failed",
			args:    args{client: ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "FAILED", Id: &id2}}), ids: ids},
			want:    []string{},
			want2:   true,
			wantErr: false,
		},
		{
			name:    "one of builds timeout",
			args:    args{client: ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "TIMED_OUT", Id: &id2}}), ids: ids},
			want:    []string{},
			want2:   true,
			wantErr: false,
		},
		{
			name:    "api error",
			args:    args{client: ReturnBatchGetBuildsMockAPI([]types.Build{}), ids: errids},
			want:    nil,
			want2:   true,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2, err := BuildStatusCheck(tt.args.client(t), tt.args.ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildStatusCheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildStatusCheck() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("buildStatusCheck() = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_RunCodeBuild(t *testing.T) {
	project := "project"
	errproject := "error"
	id := "project:12345"
	type args struct {
		client func(t *testing.T) CodeBuildAPI
		input  codebuild.StartBuildInput
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "success to start",
			args:    args{client: ReturnStartBuildMockAPI(&types.Build{Id: &id}, nil), input: codebuild.StartBuildInput{ProjectName: &project}},
			want:    id,
			wantErr: false,
		},
		{
			name:    "api error",
			args:    args{client: ReturnStartBuildMockAPI(&types.Build{Id: &id}, nil), input: codebuild.StartBuildInput{ProjectName: &errproject}},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunCodeBuild(tt.args.client(t), tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("runCodeBuild() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("runCodeBuild() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_RetryCodeBuild(t *testing.T) {
	id1 := "project:12345678"
	type args struct {
		client func(t *testing.T) CodeBuildAPI
		id     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "basic",
			args:    args{client: ReturnRetryBuildMockAPI(types.Build{Id: &id1}), id: id1},
			want:    id1,
			wantErr: false,
		},
		{
			name:    "api error",
			args:    args{client: ReturnRetryBuildMockAPI(types.Build{}), id: "error"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RetryCodeBuild(tt.args.client(t), tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("retryCodeBuild() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("retryCodeBuild() = %v, want %v", got, tt.want)
			}
		})
	}
}