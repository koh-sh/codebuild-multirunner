package cwlog

import (
	"fmt"
	"log"

	root "github.com/koh-sh/codebuild-multirunner/cmd"
	"github.com/koh-sh/codebuild-multirunner/internal/cwlog"
	mr "github.com/koh-sh/codebuild-multirunner/internal/multirunner"
	"github.com/spf13/cobra"
)

// options
var id string

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Print CodeBuild log for a single build with a provided id.",
	Long: `Print CodeBuild log for a single build with a provided id.

Only CloudWatch Logs is supported.
S3 Log is not supported`,

	Run: func(cmd *cobra.Command, args []string) {
		cbclient, err := mr.NewCodeBuildAPI()
		if err != nil {
			log.Fatal(err)
		}
		group, stream, err := cwlog.GetCloudWatchLogSetting(cbclient, id)
		if err != nil {
			log.Fatal(err)
		}
		cwlclient, err := mr.NewCloudWatchLogsAPI()
		if err != nil {
			log.Fatal(err)
		}
		// first request will be invoked without token
		token := ""
		for {
			res, err := cwlog.GetCloudWatchLogEvents(cwlclient, group, stream, token)
			if err != nil {
				log.Fatal(err)
			}
			// NextForwardToken is..
			// The token for the next set of items in the forward direction. The token expires
			// after 24 hours. If you have reached the end of the stream, it returns the same
			// token you passed in.
			if *res.NextForwardToken == token {
				break
			}
			token = *res.NextForwardToken
			for _, event := range res.Events {
				fmt.Print(*event.Message)
			}
		}
	},
}

func init() {
	root.RootCmd.AddCommand(logCmd)
	logCmd.Flags().StringVar(&id, "id", "", "CodeBuild build id for getting log")
	logCmd.MarkFlagRequired("id")
}
