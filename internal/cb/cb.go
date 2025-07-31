package cb

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/fatih/color"
	"github.com/goccy/go-yaml"
	"github.com/jinzhu/copier"
	"github.com/koh-sh/codebuild-multirunner/internal/types"
)

// interface for AWS CodeBuild API
type CodeBuildAPI interface {
	BatchGetBuilds(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error)
	StartBuild(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error)
	RetryBuild(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error)
}

// return CodeBuild api client
func NewCodeBuildAPI() (CodeBuildAPI, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	return codebuild.NewFromConfig(cfg), nil
}

// run CodeBuild Projects and return build id
func RunCodeBuild(client CodeBuildAPI, input codebuild.StartBuildInput) (string, error) {
	result, err := client.StartBuild(context.Background(), &input)
	if err != nil {
		return "", err
	}
	id := *result.Build.Id
	log.Printf("%s [STARTED]\n", id)
	return id, nil
}

// retry CodeBuild build
func RetryCodeBuild(client CodeBuildAPI, id string) (string, error) {
	input := codebuild.RetryBuildInput{Id: &id}
	result, err := client.RetryBuild(context.Background(), &input)
	if err != nil {
		return "", err
	}
	buildid := *result.Build.Id
	log.Printf("%s [STARTED]\n", buildid)
	return buildid, err
}

// read yaml config file for builds definition
// returns parsed builds (map or list) and a boolean indicating if it's the map format
func ReadConfigFile(filepath string) (any, bool, error) {
	var data map[string]any
	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, false, err
	}
	expanded := os.ExpandEnv(string(b))
	err = yaml.Unmarshal([]byte(expanded), &data)
	if err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	buildsData, ok := data["builds"]
	if !ok {
		return nil, false, fmt.Errorf("`builds` field not found in config file")
	}

	switch buildsTyped := buildsData.(type) {
	case map[string]any:
		// New map format
		parsedMap := make(map[string][]types.Build)
		buildsYAML, err := yaml.Marshal(buildsTyped) // Re-marshal to handle nested lists correctly
		if err != nil {
			return nil, true, fmt.Errorf("failed to re-marshal map builds: %w", err)
		}
		err = yaml.Unmarshal(buildsYAML, &parsedMap)
		if err != nil {
			return nil, true, fmt.Errorf("failed to unmarshal map builds into target type: %w", err)
		}
		return parsedMap, true, nil
	case []any:
		// Legacy list format
		fmt.Fprintf(os.Stderr, "\n⚠️  WARNING: List format for 'builds' is deprecated. Please migrate to map format.\n\n")
		parsedList := []types.Build{}
		buildsYAML, err := yaml.Marshal(buildsTyped) // Re-marshal to handle list items correctly
		if err != nil {
			return nil, false, fmt.Errorf("failed to re-marshal list builds: %w", err)
		}
		err = yaml.Unmarshal(buildsYAML, &parsedList)
		if err != nil {
			return nil, false, fmt.Errorf("failed to unmarshal list builds into target type: %w", err)
		}
		return parsedList, false, nil
	default:
		return nil, false, fmt.Errorf("unexpected type for 'builds' field: %T", buildsData)
	}
}

// dump read config with environment variables inserted
func DumpConfig(configfile string) (string, error) {
	// Use ReadConfigFile to ensure deprecation warnings are shown
	builds, _, err := ReadConfigFile(configfile)
	if err != nil {
		return "", err
	}

	// Reconstruct the config structure for dumping
	configData := map[string]any{"builds": builds}

	// Marshal with options for pretty printing
	d, err := yaml.MarshalWithOptions(&configData, yaml.Indent(4), yaml.IndentSequence(true))
	if err != nil {
		return "", fmt.Errorf("failed to marshal config for dump: %w", err)
	}
	return string(d), nil
}

// copy configuration read from yaml to codebuild.StartBuildInput
func ConvertBuildConfigToStartBuildInput(build types.Build) (codebuild.StartBuildInput, error) {
	startbuildinput := codebuild.StartBuildInput{}
	err := copier.CopyWithOption(&startbuildinput, build, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		return startbuildinput, err
	}
	return startbuildinput, nil
}

// wait and check status of builds and return if any build failed
func WaitAndCheckBuildStatus(client CodeBuildAPI, ids []string, pollsec int) (bool, error) {
	var err error
	failed := false
	hasfailed := false
	for {
		// break if all builds end
		if len(ids) == 0 {
			return hasfailed, nil
		}
		time.Sleep(time.Duration(pollsec) * time.Second)
		ids, failed, err = buildStatusCheck(client, ids)
		if err != nil {
			return false, err
		}
		if failed {
			hasfailed = true
		}
	}
}

// check builds status and return ongoing build ids
func buildStatusCheck(client CodeBuildAPI, ids []string) ([]string, bool, error) {
	inprogressids := []string{}
	hasfailedbuild := false
	input := codebuild.BatchGetBuildsInput{Ids: ids}
	result, err := client.BatchGetBuilds(context.Background(), &input)
	if err != nil {
		return nil, true, err
	}
	for _, v := range result.Builds {
		log.Printf("%s [%s]\n", *v.Id, coloredString(string(v.BuildStatus)))
		if v.BuildStatus == "IN_PROGRESS" {
			inprogressids = append(inprogressids, *v.Id)
		} else if v.BuildStatus != "SUCCEEDED" {
			hasfailedbuild = true
		}
	}
	return inprogressids, hasfailedbuild, nil
}

// return colored string for each CodeBuild statuses
func coloredString(status string) string {
	switch status {
	case "SUCCEEDED":
		return color.GreenString(status)
	case "IN_PROGRESS":
		return color.BlueString(status)
	default:
		return color.RedString(status)
	}
}

// FilterBuildsByTarget filters the builds based on the provided targets.
// It returns a list of builds to run and an error if any target is invalid or not found.
func FilterBuildsByTarget(parsedBuilds any, isMapFormat bool, targets []string) ([]types.Build, error) {
	var buildsToRun []types.Build

	if isMapFormat {
		groupedBuilds := parsedBuilds.(map[string][]types.Build)
		if len(targets) == 0 {
			// Run all builds from all groups if no targets specified
			for _, groupBuilds := range groupedBuilds {
				buildsToRun = append(buildsToRun, groupBuilds...)
			}
		} else {
			// Run builds only from specified target groups
			foundTargets := make(map[string]bool)
			for _, targetGroup := range targets {
				if groupBuilds, ok := groupedBuilds[targetGroup]; ok {
					buildsToRun = append(buildsToRun, groupBuilds...)
					foundTargets[targetGroup] = true
				} else {
					return nil, fmt.Errorf("targets group '%s' not found in config file", targetGroup)
				}
			}
			// Check if any specified targets were not found
			allTargetsFound := true
			for _, t := range targets {
				if !foundTargets[t] {
					allTargetsFound = false
				}
			}
			if !allTargetsFound {
				return nil, fmt.Errorf("one or more specified targets groups were not found")
			}
		}
	} else {
		// For list format, targets option is not supported
		if len(targets) > 0 {
			return nil, fmt.Errorf("--targets option is only available for the map format configuration file")
		}
		buildsToRun = parsedBuilds.([]types.Build)
	}

	return buildsToRun, nil
}
