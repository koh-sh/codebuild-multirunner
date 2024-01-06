package cmd

import (
	"log"
	"os"

	cb "github.com/koh-sh/codebuild-multirunner/internal/codebuild"
	"github.com/spf13/cobra"
)

// retryCmd represents the retry command
var retryCmd = &cobra.Command{
	Use:   "retry",
	Short: "retry CodeBuild build with a provided id",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := cb.NewCodeBuildAPI()
		if err != nil {
			log.Fatal(err)
		}
		buildid, err := cb.RetryCodeBuild(client, id)
		if err != nil {
			log.Fatal(err)
		}
		ids := []string{buildid}
		failed := false
		// early return if --no-wait option set
		if nowait {
			return
		}
		failed, err = cb.WaitAndCheckBuildStatus(client, ids, pollsec)
		if err != nil {
			log.Fatal(err)
		}
		if failed {
			os.Exit(2)
		}
	},
}

func init() {
	rootCmd.AddCommand(retryCmd)
	retryCmd.Flags().BoolVar(&nowait, "no-wait", false, "specify if you don't need to follow builds status")
	retryCmd.Flags().IntVar(&pollsec, "polling-span", 60, "polling span in second for builds status check")
	retryCmd.Flags().StringVar(&id, "id", "", "CodeBuild build id for retry")
	retryCmd.MarkFlagRequired("id")
}
