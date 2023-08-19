package cmd

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/fatih/color"
	root "github.com/koh-sh/codebuild-multirunner/cmd"
	"github.com/spf13/cobra"
)

// options
var id string
var nowait bool
var pollsec int

// retryCmd represents the retry command
var retryCmd = &cobra.Command{
	Use:   "retry",
	Short: "retry CodeBuild build with a provided id",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := root.NewCodeBuildAPI()
		if err != nil {
			log.Fatal(err)
		}
		buildid, err := retryCodeBuild(client, id)
		if err != nil {
			log.Fatal(err)
		}
		ids := []string{buildid}
		hasfailedbuild := false
		// early return if --no-wait option set
		if nowait {
			return
		}
		for i := 0; ; i++ {
			// break if all builds end
			if len(ids) == 0 {
				break
			}
			time.Sleep(time.Duration(pollsec) * time.Second)
			failed := false
			ids, failed = buildStatusCheck(client, ids)
			if failed {
				hasfailedbuild = true
			}
		}
		if hasfailedbuild {
			os.Exit(2)
		}
	},
}

func init() {
	root.RootCmd.AddCommand(retryCmd)
	retryCmd.Flags().BoolVar(&nowait, "no-wait", false, "specify if you don't need to follow builds status")
	retryCmd.Flags().IntVar(&pollsec, "polling-span", 60, "polling span in second for builds status check")
	retryCmd.Flags().StringVar(&id, "id", "", "CodeBuild build id for retry")
	retryCmd.MarkFlagRequired("id")
}

// retry CodeBuild build
func retryCodeBuild(client root.CodeBuildAPI, id string) (string, error) {
	input := codebuild.RetryBuildInput{Id: &id}
	result, err := client.RetryBuild(context.TODO(), &input)
	if err != nil {
		return "", err
	}
	buildid := *result.Build.Id
	log.Printf("%s [STARTED]\n", buildid)
	return buildid, err
}

// check builds status and return ongoing build ids
func buildStatusCheck(client root.CodeBuildAPI, ids []string) ([]string, bool) {
	inprogressids := []string{}
	hasfailedbuild := false
	input := codebuild.BatchGetBuildsInput{Ids: ids}
	result, err := client.BatchGetBuilds(context.TODO(), &input)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range result.Builds {
		log.Printf("%s [%s]\n", *v.Id, coloredString(string(v.BuildStatus)))
		if v.BuildStatus == "IN_PROGRESS" {
			inprogressids = append(inprogressids, *v.Id)
		} else if v.BuildStatus != "SUCCEEDED" {
			hasfailedbuild = true
		}
	}
	return inprogressids, hasfailedbuild
}

// return colored string for each CodeBuild statuses
func coloredString(status string) string {
	if status == "SUCCEEDED" {
		return color.GreenString(status)
	} else if status == "IN_PROGRESS" {
		return color.BlueString(status)
	} else {
		return color.RedString(status)
	}
}
