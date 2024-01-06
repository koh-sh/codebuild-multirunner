package multirunner

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/fatih/color"
	cmt "github.com/koh-sh/codebuild-multirunner/internal/types"
)

func Test_readConfigFile(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		want    cmt.BuildConfig
		wantErr bool
	}{
		{
			name: "basic",
			args: args{"testdata/_test.yaml"},
			want: cmt.BuildConfig{
				Builds: []cmt.Build{
					{ProjectName: "testproject", SourceVersion: "chore/test"},
					{ProjectName: "testproject2"},
				},
			},
			wantErr: false,
		},
		{
			name: "environment variable",
			args: args{"testdata/_test2.yaml"},
			want: cmt.BuildConfig{
				Builds: []cmt.Build{
					{ProjectName: "testproject3", SourceVersion: "chore/test"},
					{ProjectName: "testproject2"},
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid yaml file",
			args:    args{"testdata/_test3.yaml"},
			want:    cmt.BuildConfig{},
			wantErr: true,
		},
		{
			name:    "file not found",
			args:    args{"testdata/_testxxx.yaml"},
			want:    cmt.BuildConfig{},
			wantErr: true,
		},
	}
	t.Setenv("TEST_ENV", "testproject3") // setting environment variable for test case 2
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadConfigFile(tt.args.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("readConfigFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readConfigFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ColoredString(t *testing.T) {
	type args struct {
		status string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "SUCCEEDED",
			args: args{status: "SUCCEEDED"},
			want: color.GreenString("SUCCEEDED"),
		},
		{
			name: "IN_PROGRESS",
			args: args{status: "IN_PROGRESS"},
			want: color.BlueString("IN_PROGRESS"),
		},
		{
			name: "FAILED",
			args: args{status: "FAILED"},
			want: color.RedString("FAILED"),
		},
		{
			name: "TIMED_OUT",
			args: args{status: "TIMED_OUT"},
			want: color.RedString("TIMED_OUT"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ColoredString(tt.args.status); got != tt.want {
				t.Errorf("coloredString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_DumpConfig(t *testing.T) {
	wantyaml := `builds:
    - projectName: testproject
      sourceVersion: chore/test
    - projectName: testproject2
`
	type args struct {
		bc cmt.BuildConfig
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				cmt.BuildConfig{
					Builds: []cmt.Build{
						{ProjectName: "testproject", SourceVersion: "chore/test"},
						{ProjectName: "testproject2"},
					},
				},
			},
			want:    wantyaml,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DumpConfig(tt.args.bc)
			if (err != nil) != tt.wantErr {
				t.Errorf("dumpConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("dumpConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ConvertBuildConfigToStartBuildInput(t *testing.T) {
	type args struct {
		build cmt.Build
	}
	tests := []struct {
		name    string
		args    args
		want    codebuild.StartBuildInput
		wantErr bool
	}{
		{
			name:    "basic",
			args:    args{cmt.Build{}},
			want:    codebuild.StartBuildInput{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertBuildConfigToStartBuildInput(tt.args.build)
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
