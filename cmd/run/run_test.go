package cmd

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/koh-sh/codebuild-multirunner/common"
)

func Test_convertBuildConfigToStartBuildInput(t *testing.T) {
	type args struct {
		build common.Build
	}
	tests := []struct {
		name string
		args args
		want codebuild.StartBuildInput
	}{
		{
			name: "basic",
			args: args{common.Build{}},
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

func Test_runCodeBuild(t *testing.T) {
	project := "project"
	id := "project:12345"
	type args struct {
		client func(t *testing.T) common.CodeBuildAPI
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
			args:    args{client: common.ReturnStartBuildMockAPI(&types.Build{Id: &id}, nil), input: codebuild.StartBuildInput{ProjectName: &project}},
			want:    id,
			wantErr: false,
		},
		{
			name:    "fail to start",
			args:    args{client: common.ReturnStartBuildMockAPI(&types.Build{Id: &id}, errors.New("fail to run")), input: codebuild.StartBuildInput{ProjectName: &project}},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runCodeBuild(tt.args.client(t), tt.args.input)
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
