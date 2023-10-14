package cmd

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	mr "github.com/koh-sh/codebuild-multirunner/internal/multirunner"
)

func Test_convertBuildConfigToStartBuildInput(t *testing.T) {
	type args struct {
		build mr.Build
	}
	tests := []struct {
		name    string
		args    args
		want    codebuild.StartBuildInput
		wantErr bool
	}{
		{
			name:    "basic",
			args:    args{mr.Build{}},
			want:    codebuild.StartBuildInput{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertBuildConfigToStartBuildInput(tt.args.build)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertBuildConfigToStartBuildInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertBuildConfigToStartBuildInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_runCodeBuild(t *testing.T) {
	project := "project"
	errproject := "error"
	id := "project:12345"
	type args struct {
		client func(t *testing.T) mr.CodeBuildAPI
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
			args:    args{client: mr.ReturnStartBuildMockAPI(&types.Build{Id: &id}, nil), input: codebuild.StartBuildInput{ProjectName: &project}},
			want:    id,
			wantErr: false,
		},
		{
			name:    "api error",
			args:    args{client: mr.ReturnStartBuildMockAPI(&types.Build{Id: &id}, nil), input: codebuild.StartBuildInput{ProjectName: &errproject}},
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
