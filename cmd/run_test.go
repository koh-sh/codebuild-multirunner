package cmd

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/fatih/color"
)

// mock api
type BatchGetBuildsMockAPI func(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error)

// StartBuild implements CodeBuildAPI.
func (m BatchGetBuildsMockAPI) StartBuild(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error) {
	panic("unimplemented")
}

// return m(ctx, params, optFns...)
func (m BatchGetBuildsMockAPI) BatchGetBuilds(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
	// return params
	return m(ctx, params, optFns...)
}

// return mock function
func ReturnBatchGetBuildsMockAPI(builds []types.Build) func(t *testing.T) CodeBuildAPI {
	mock := func(t *testing.T) CodeBuildAPI {
		return BatchGetBuildsMockAPI(func(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
			t.Helper()
			if len(params.Ids) == 0 {
				t.Fatal("Ids must have at least one")
			}
			return &codebuild.BatchGetBuildsOutput{
				Builds:         builds,
				BuildsNotFound: []string{},
				ResultMetadata: middleware.Metadata{},
			}, nil
		})
	}
	return mock
}

func Test_convertBuildConfigToStartBuildInput(t *testing.T) {
	type args struct {
		build Build
	}
	tests := []struct {
		name string
		args args
		want codebuild.StartBuildInput
	}{
		{name: "basic",
			args: args{Build{}},
			want: codebuild.StartBuildInput{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertBuildConfigToStartBuildInput(tt.args.build); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertBuildConfigToStartBuildInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_coloredString(t *testing.T) {
	type args struct {
		status string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "SUCCEEDED",
			args: args{status: "SUCCEEDED"},
			want: color.GreenString("SUCCEEDED")},
		{name: "IN_PROGRESS",
			args: args{status: "IN_PROGRESS"},
			want: color.BlueString("IN_PROGRESS")},
		{name: "FAILED",
			args: args{status: "FAILED"},
			want: color.RedString("FAILED")},
		{name: "TIMED_OUT",
			args: args{status: "TIMED_OUT"},
			want: color.RedString("TIMED_OUT")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := coloredString(tt.args.status); got != tt.want {
				t.Errorf("coloredString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildStatusCheck(t *testing.T) {
	var id1 = "project:12345678"
	var id2 = "project2:87654321"
	var ids = []string{id1, id2}

	type args struct {
		client func(t *testing.T) CodeBuildAPI
		ids    []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "all builds ended",
			args: args{client: ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "FAILED", Id: &id2}}), ids: ids},
			want: []string{},
		},
		{name: "one builds in progress",
			args: args{client: ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "IN_PROGRESS", Id: &id2}}), ids: ids},
			want: []string{id2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildStatusCheck(tt.args.client(t), tt.args.ids); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildStatusCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}
