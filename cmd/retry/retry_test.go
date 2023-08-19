package cmd

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/fatih/color"
	root "github.com/koh-sh/codebuild-multirunner/cmd"
)

func Test_retryCodeBuild(t *testing.T) {
	var id1 = "project:12345678"
	type args struct {
		client func(t *testing.T) root.CodeBuildAPI
		id     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "basic",
			args:    args{client: root.ReturnRetryBuildMockAPI(types.Build{Id: &id1}), id: id1},
			want:    id1,
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := retryCodeBuild(tt.args.client(t), tt.args.id)
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
		client func(t *testing.T) root.CodeBuildAPI
		ids    []string
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want2 bool
	}{
		{name: "all builds ended",
			args:  args{client: root.ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "SUCCEEDED", Id: &id2}}), ids: ids},
			want:  []string{},
			want2: false,
		},
		{name: "one builds in progress",
			args:  args{client: root.ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "IN_PROGRESS", Id: &id2}}), ids: ids},
			want:  []string{id2},
			want2: false,
		},
		{name: "one of builds failed",
			args:  args{client: root.ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "FAILED", Id: &id2}}), ids: ids},
			want:  []string{},
			want2: true,
		},
		{name: "one of builds timeout",
			args:  args{client: root.ReturnBatchGetBuildsMockAPI([]types.Build{{BuildStatus: "SUCCEEDED", Id: &id1}, {BuildStatus: "TIMED_OUT", Id: &id2}}), ids: ids},
			want:  []string{},
			want2: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2 := buildStatusCheck(tt.args.client(t), tt.args.ids)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildStatusCheck() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("buildStatusCheck() = %v, want %v", got2, tt.want2)
			}
		})
	}
}
