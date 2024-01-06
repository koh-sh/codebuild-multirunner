package codebuild

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	mr "github.com/koh-sh/codebuild-multirunner/internal/multirunner"
)

// interface for AWS CodeBuild API
type CodeBuildAPI interface {
	BatchGetBuilds(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error)
	StartBuild(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error)
	RetryBuild(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error)
}

// return CodeBuild api client
func NewCodeBuildAPI() (CodeBuildAPI, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	return codebuild.NewFromConfig(cfg), nil
}

// check builds status and return ongoing build ids
func BuildStatusCheck(client CodeBuildAPI, ids []string) ([]string, bool, error) {
	inprogressids := []string{}
	hasfailedbuild := false
	input := codebuild.BatchGetBuildsInput{Ids: ids}
	result, err := client.BatchGetBuilds(context.TODO(), &input)
	if err != nil {
		return nil, true, err
	}
	for _, v := range result.Builds {
		log.Printf("%s [%s]\n", *v.Id, mr.ColoredString(string(v.BuildStatus)))
		if v.BuildStatus == "IN_PROGRESS" {
			inprogressids = append(inprogressids, *v.Id)
		} else if v.BuildStatus != "SUCCEEDED" {
			hasfailedbuild = true
		}
	}
	return inprogressids, hasfailedbuild, nil
}

// run CodeBuild Projects and return build id
func RunCodeBuild(client CodeBuildAPI, input codebuild.StartBuildInput) (string, error) {
	result, err := client.StartBuild(context.TODO(), &input)
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
	result, err := client.RetryBuild(context.TODO(), &input)
	if err != nil {
		return "", err
	}
	buildid := *result.Build.Id
	log.Printf("%s [STARTED]\n", buildid)
	return buildid, err
}
