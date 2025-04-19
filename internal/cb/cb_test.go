package cb

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
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
		name            string
		args            args
		want            any
		wantErr         bool
		wantErrContains string // Optional: check if error message contains this string
	}{
		{
			name: "basic (list format)",
			args: args{"testdata/_test.yaml"},
			want: []cmt.Build{
				{ProjectName: "testproject", SourceVersion: "chore/test"},
				{ProjectName: "testproject2"},
			},
			wantErr: false,
		},
		{
			name: "environment variable (list format)",
			args: args{"testdata/_test2.yaml"},
			want: []cmt.Build{
				{ProjectName: "testproject3", SourceVersion: "chore/test"},
				{ProjectName: "testproject2"},
			},
			wantErr: false,
		},
		{
			name: "map format",
			args: args{"testdata/_test_map.yaml"},
			want: map[string][]cmt.Build{
				"group1": {
					{ProjectName: "proj-a"},
				},
				"group2": {
					{ProjectName: "proj-b", SourceVersion: "develop"},
					{ProjectName: "proj-c"},
				},
			},
			wantErr: false,
		},
		{
			name:            "invalid yaml file",
			args:            args{"testdata/_test3.yaml"},
			want:            nil,
			wantErr:         true,
			wantErrContains: "failed to unmarshal yaml",
		},
		{
			name:    "file not found",
			args:    args{"testdata/_testxxx.yaml"},
			want:    nil,
			wantErr: true,
		},
		{
			name:            "missing builds field",
			args:            args{"testdata/_test_missing_builds.yaml"},
			want:            nil,
			wantErr:         true,
			wantErrContains: "`builds` field not found",
		},
		{
			name:            "invalid builds field type",
			args:            args{"testdata/_test_invalid_builds_type.yaml"},
			want:            nil,
			wantErr:         true,
			wantErrContains: "unexpected type for 'builds' field",
		},
	}
	t.Setenv("TEST_ENV", "testproject3")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := ReadConfigFile(tt.args.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadConfigFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("ReadConfigFile() error = %v, wantErr containing %q", err, tt.wantErrContains)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadConfigFile() got = %#v, want %#v", got, tt.want)
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
		name            string
		args            args
		want            string
		wantErr         bool
		wantErrContains string // Optional: check if error message contains this string
	}{
		{
			name:    "basic",
			args:    args{"testdata/_test.yaml"},
			want:    wantyaml,
			wantErr: false,
		},
		{
			name:            "file not found",
			args:            args{"testdata/_test_dump_notfound.yaml"},
			want:            "",
			wantErr:         true,
			wantErrContains: "failed to read config file for dump",
		},
		{
			name:            "invalid yaml syntax",
			args:            args{"testdata/_test_invalid_syntax_dump.yaml"},
			want:            "",
			wantErr:         true,
			wantErrContains: "failed to unmarshal yaml for dump",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DumpConfig(tt.args.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("DumpConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("DumpConfig() error = %v, wantErr containing %q", err, tt.wantErrContains)
			}
			if got != tt.want {
				t.Errorf("DumpConfig() got = %v, want %v", got, tt.want)
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

func TestFilterBuildsByTarget(t *testing.T) {
	mapBuilds := map[string][]cmt.Build{
		"group1": {
			{ProjectName: "proj-a"},
		},
		"group2": {
			{ProjectName: "proj-b", SourceVersion: "develop"},
			{ProjectName: "proj-c"},
		},
		"emptyGroup": {},
	}

	listBuilds := []cmt.Build{
		{ProjectName: "testproject", SourceVersion: "chore/test"},
		{ProjectName: "testproject2"},
	}

	tests := []struct {
		name         string
		parsedBuilds any
		isMapFormat  bool
		targets      []string
		want         []cmt.Build
		wantErr      bool
		// Flag to indicate order-independent comparison should be used
		orderIndependent bool
	}{
		{
			name:         "Map format, no targets",
			parsedBuilds: mapBuilds,
			isMapFormat:  true,
			targets:      []string{},
			want: []cmt.Build{
				{ProjectName: "proj-a"},
				{ProjectName: "proj-b", SourceVersion: "develop"},
				{ProjectName: "proj-c"},
			},
			wantErr:          false,
			orderIndependent: true,
		},
		{
			name:         "Map format, one target",
			parsedBuilds: mapBuilds,
			isMapFormat:  true,
			targets:      []string{"group1"},
			want: []cmt.Build{
				{ProjectName: "proj-a"},
			},
			wantErr: false,
		},
		{
			name:         "Map format, multiple targets",
			parsedBuilds: mapBuilds,
			isMapFormat:  true,
			targets:      []string{"group1", "group2"},
			want: []cmt.Build{
				{ProjectName: "proj-a"},
				{ProjectName: "proj-b", SourceVersion: "develop"},
				{ProjectName: "proj-c"},
			},
			wantErr:          false,
			orderIndependent: true,
		},
		{
			name:         "Map format, target not found",
			parsedBuilds: mapBuilds,
			isMapFormat:  true,
			targets:      []string{"group3"},
			want:         nil,
			wantErr:      true,
		},
		{
			name:         "Map format, some targets not found",
			parsedBuilds: mapBuilds,
			isMapFormat:  true,
			targets:      []string{"group1", "group3"},
			want:         nil, // Expect error, so specific builds don't matter
			wantErr:      true,
		},
		{
			name:         "Map format, target is empty group",
			parsedBuilds: mapBuilds,
			isMapFormat:  true,
			targets:      []string{"emptyGroup"},
			want:         []cmt.Build{}, // Expect empty slice, no error
			wantErr:      false,
		},
		{
			name:         "List format, no targets",
			parsedBuilds: listBuilds,
			isMapFormat:  false,
			targets:      []string{},
			want: []cmt.Build{
				{ProjectName: "testproject", SourceVersion: "chore/test"},
				{ProjectName: "testproject2"},
			},
			wantErr: false,
		},
		{
			name:         "List format, targets specified (should error)",
			parsedBuilds: listBuilds,
			isMapFormat:  false,
			targets:      []string{"group1"},
			want:         nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilterBuildsByTarget(tt.parsedBuilds, tt.isMapFormat, tt.targets)

			// Check error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterBuildsByTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expect an error, don't check the result further
			if tt.wantErr {
				return
			}

			// Both nil and empty slices represent "no builds" - consider them equal
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			// For cases where order doesn't matter, use content-based comparison
			if tt.orderIndependent {
				// Check for same length first
				if len(got) != len(tt.want) {
					t.Errorf("FilterBuildsByTarget() got %d builds, want %d builds", len(got), len(tt.want))
					return
				}

				// Create lookup map by ProjectName
				wantBuilds := make(map[string]cmt.Build)
				for _, build := range tt.want {
					wantBuilds[build.ProjectName] = build
				}

				// Check each build in the result
				for _, gotBuild := range got {
					wantBuild, exists := wantBuilds[gotBuild.ProjectName]
					if !exists {
						t.Errorf("FilterBuildsByTarget() unexpected project %q in result", gotBuild.ProjectName)
						continue
					}

					// Compare other fields
					if gotBuild.SourceVersion != wantBuild.SourceVersion {
						t.Errorf("FilterBuildsByTarget() project %q has SourceVersion = %q, want %q",
							gotBuild.ProjectName, gotBuild.SourceVersion, wantBuild.SourceVersion)
					}
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				// For cases where order matters, use exact DeepEqual
				t.Errorf("FilterBuildsByTarget() got = %v, want %v", got, tt.want)
			}
		})
	}
}
