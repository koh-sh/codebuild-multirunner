package cb

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/fatih/color"
	"github.com/jinzhu/copier"
	"github.com/koh-sh/codebuild-multirunner/internal/types"
	"gopkg.in/yaml.v3"
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

// dump read config with environment variables inserted
func DumpConfig(configfile string) (string, error) {
	bc, err := ReadConfigFile(configfile)
	if err != nil {
		log.Fatal(err)
	}
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
