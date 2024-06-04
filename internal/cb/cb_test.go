package cb

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/fatih/color"
	cmt "github.com/koh-sh/codebuild-multirunner/internal/types"
)

// mock api for StartBuild
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

func Test_buildStatusCheck(t *testing.T) {
	mockCodeBuildAPI := NewMockCodeBuildAPI(
		nil,
		func(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
			if params.Ids[0] == "error:12345678" {
				return nil, errors.New("batch get builds error")
			}
			builds := make([]types.Build, len(params.Ids))
			for i, id := range params.Ids {
				var status types.StatusType
				switch id {
				case "project2:in-progress":
					status = types.StatusTypeInProgress
				case "project3:failed":
					status = types.StatusTypeFailed
				case "project4:timeout":
					status = types.StatusTypeTimedOut
				default:
					status = types.StatusTypeSucceeded
				}
				builds[i] = types.Build{
					Id:          &id,
					BuildStatus: status,
				}
			}
			return &codebuild.BatchGetBuildsOutput{
				Builds: builds,
			}, nil
		},
		nil,
	)
	tests := []struct {
		name    string
		ids     []string
		want    []string
		want2   bool
		wantErr bool
	}{
		{
			name:    "all builds ended",
			ids:     []string{"project1:12345678", "project2:87654321"},
			want:    []string{},
			want2:   false,
			wantErr: false,
		},
		{
			name:    "one builds in progress",
			ids:     []string{"project1:12345678", "project2:in-progress"},
			want:    []string{"project2:in-progress"},
			want2:   false,
			wantErr: false,
		},
		{
			name:    "one of builds failed",
			ids:     []string{"project1:12345678", "project3:failed"},
			want:    []string{},
			want2:   true,
			wantErr: false,
		},
		{
			name:    "one of builds timeout",
			ids:     []string{"project1:12345678", "project4:timeout"},
			want:    []string{},
			want2:   true,
			wantErr: false,
		},
		{
			name:    "api error",
			ids:     []string{"error:12345678"},
			want:    nil,
			want2:   true,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2, err := buildStatusCheck(mockCodeBuildAPI, tt.ids)
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
	mockCodeBuildAPI := NewMockCodeBuildAPI(
		func(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error) {
			if *params.ProjectName == "error" {
				return nil, errors.New("start build error")
			}
			buildID := fmt.Sprintf("%s:12345678", *params.ProjectName)
			return &codebuild.StartBuildOutput{
				Build: &types.Build{
					Id: &buildID,
				},
			}, nil
		},
		nil,
		nil,
	)
	tests := []struct {
		name        string
		projectName string
		want        string
		wantErr     bool
	}{
		{
			name:        "success to start",
			projectName: "project1",
			want:        "project1:12345678",
			wantErr:     false,
		},
		{
			name:        "api error",
			projectName: "error",
			want:        "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := codebuild.StartBuildInput{
				ProjectName: &tt.projectName,
			}
			got, err := RunCodeBuild(mockCodeBuildAPI, input)
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
	mockCodeBuildAPI := NewMockCodeBuildAPI(
		nil,
		nil,
		func(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error) {
			if *params.Id == "error" {
				return nil, errors.New("retry build error")
			}
			return &codebuild.RetryBuildOutput{
				Build: &types.Build{
					Id:          params.Id,
					BuildStatus: types.StatusTypeSucceeded,
				},
			}, nil
		},
	)
	tests := []struct {
		name    string
		id      string
		want    string
		wantErr bool
	}{
		{
			name:    "basic",
			id:      "project:12345678",
			want:    "project:12345678",
			wantErr: false,
		},
		{
			name:    "api error",
			id:      "error",
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RetryCodeBuild(mockCodeBuildAPI, tt.id)
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

func Test_ReadConfigFile(t *testing.T) {
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

func Test_coloredString(t *testing.T) {
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
			if got := coloredString(tt.args.status); got != tt.want {
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
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "basic",
			args:    args{"testdata/_test.yaml"},
			want:    wantyaml,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DumpConfig(tt.args.filepath)
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

func TestWaitAndCheckBuildStatus(t *testing.T) {
	mockCodeBuildAPI := NewMockCodeBuildAPI(
		nil,
		func(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
			if params.Ids[0] == "error:12345678" {
				return nil, errors.New("batch get builds error")
			}
			builds := make([]types.Build, len(params.Ids))
			for i, id := range params.Ids {
				var status types.StatusType
				switch id {
				case "project2:in-progress":
					status = types.StatusTypeInProgress
				case "project3:failed":
					status = types.StatusTypeFailed
				case "project4:timeout":
					status = types.StatusTypeTimedOut
				default:
					status = types.StatusTypeSucceeded
				}
				builds[i] = types.Build{
					Id:          &id,
					BuildStatus: status,
				}
			}
			return &codebuild.BatchGetBuildsOutput{
				Builds: builds,
			}, nil
		},
		nil,
	)
	tests := []struct {
		name    string
		ids     []string
		pollsec int
		want    bool
		wantErr bool
	}{
		{
			name:    "all build successed",
			ids:     []string{"project:12345678", "project2:22345678"},
			pollsec: 0,
			want:    false,
			wantErr: false,
		},
		{
			name:    "one of builds failed",
			ids:     []string{"project:12345678", "project3:failed"},
			pollsec: 0,
			want:    true,
			wantErr: false,
		},
		{
			name:    "api error",
			ids:     []string{"error:12345678"},
			pollsec: 0,
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WaitAndCheckBuildStatus(mockCodeBuildAPI, tt.ids, tt.pollsec)
			if (err != nil) != tt.wantErr {
				t.Errorf("WaitAndCheckBuildStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WaitAndCheckBuildStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
