package common

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/fatih/color"
)

func Test_readConfigFile(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name string
		args args
		want BuildConfig
	}{
		{name: "basic",
			args: args{"testdata/_test.yaml"},
			want: BuildConfig{[]Build{
				{ProjectName: "testproject", SourceVersion: "chore/test"},
				{ProjectName: "testproject2"},
			},
			},
		},
		{name: "environment variable",
			args: args{"testdata/_test2.yaml"},
			want: BuildConfig{[]Build{
				{ProjectName: "testproject3", SourceVersion: "chore/test"},
				{ProjectName: "testproject2"},
			},
			},
		},
	}
	t.Setenv("TEST_ENV", "testproject3") // setting environment variable for test case 2
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReadConfigFile(tt.args.filepath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readConfigFile() = %v, want %v", got, tt.want)
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

func Test_BuildStatusCheck(t *testing.T) {
	var id1 = "project:12345678"
	var id2 = "project2:87654321"
	var ids = []string{id1, id2}

	type args struct {
		client func(t *testing.T) CodeBuildAPI
		ids    []string
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want2 bool
	}{
		{name: "all builds ended",
			args:  args{client: ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "SUCCEEDED", Id: &id2}}), ids: ids},
			want:  []string{},
			want2: false,
		},
		{name: "one builds in progress",
			args:  args{client: ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "IN_PROGRESS", Id: &id2}}), ids: ids},
			want:  []string{id2},
			want2: false,
		},
		{name: "one of builds failed",
			args:  args{client: ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "FAILED", Id: &id2}}), ids: ids},
			want:  []string{},
			want2: true,
		},
		{name: "one of builds timeout",
			args:  args{client: ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "TIMED_OUT", Id: &id2}}), ids: ids},
			want:  []string{},
			want2: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2 := BuildStatusCheck(tt.args.client(t), tt.args.ids)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildStatusCheck() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("buildStatusCheck() = %v, want %v", got2, tt.want2)
			}
		})
	}
}
