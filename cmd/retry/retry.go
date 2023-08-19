package cmd

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	root "github.com/koh-sh/codebuild-multirunner/cmd"
	"github.com/koh-sh/codebuild-multirunner/common"
	"github.com/spf13/cobra"
)

// options
var (
	id      string
	nowait  bool
	pollsec int
)

// retryCmd represents the retry command
var retryCmd = &cobra.Command{
	Use:   "retry",
	Short: "retry CodeBuild build with a provided id",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := common.NewCodeBuildAPI()
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
			ids, failed = common.BuildStatusCheck(client, ids)
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
func retryCodeBuild(client common.CodeBuildAPI, id string) (string, error) {
	input := codebuild.RetryBuildInput{Id: &id}
	result, err := client.RetryBuild(context.TODO(), &input)
	if err != nil {
		return "", err
	}
	buildid := *result.Build.Id
	log.Printf("%s [STARTED]\n", buildid)
	return buildid, err
}
