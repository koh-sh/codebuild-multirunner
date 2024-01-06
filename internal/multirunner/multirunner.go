package multirunner

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/fatih/color"
	"github.com/jinzhu/copier"
	"github.com/koh-sh/codebuild-multirunner/internal/types"
	"gopkg.in/yaml.v3"
)

// read yaml config file for builds definition
func ReadConfigFile(filepath string) (types.BuildConfig, error) {
	bc := types.BuildConfig{}
	b, err := os.ReadFile(filepath)
	if err != nil {
		return bc, err
	}
	expanded := os.ExpandEnv(string(b))
	err = yaml.Unmarshal([]byte(expanded), &bc)
	if err != nil {
		return bc, err
	}
	return bc, nil
}

// return colored string for each CodeBuild statuses
func ColoredString(status string) string {
	switch status {
	case "SUCCEEDED":
		return color.GreenString(status)
	case "IN_PROGRESS":
		return color.BlueString(status)
	default:
		return color.RedString(status)
	}
}

// dump read config with environment variables inserted
func DumpConfig(bc types.BuildConfig) (string, error) {
	d, err := yaml.Marshal(&bc)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

// copy configration read from yaml to codebuild.StartBuildInput
func ConvertBuildConfigToStartBuildInput(build types.Build) (codebuild.StartBuildInput, error) {
	startbuildinput := codebuild.StartBuildInput{}
	err := copier.CopyWithOption(&startbuildinput, build, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		return startbuildinput, err
	}
	return startbuildinput, nil
}
