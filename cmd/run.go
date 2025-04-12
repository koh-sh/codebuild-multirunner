package cmd

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/koh-sh/codebuild-multirunner/internal/cb"
	"github.com/koh-sh/codebuild-multirunner/internal/types"
	"github.com/spf13/cobra"
)

var targets []string

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run CodeBuild projects based on YAML",
	Run: func(cmd *cobra.Command, args []string) {
		parsedBuilds, isMapFormat, err := cb.ReadConfigFile(configfile)
		if err != nil {
			log.Fatalf("Error reading config file: %v\n", err)
		}

		client, err := cb.NewCodeBuildAPI()
		if err != nil {
			log.Fatal(err)
		}

		// Determine builds to run using the new function in internal/cb
		buildsToRun, err := cb.FilterBuildsByTarget(parsedBuilds, isMapFormat, targets)
		if err != nil {
			log.Fatalf("Error filtering builds: %v\n", err)
		}

		// Run specified codebuild projects in parallel
		var wg sync.WaitGroup
		idsChan := make(chan string, len(buildsToRun))
		errChan := make(chan error, len(buildsToRun))

		for _, build := range buildsToRun {
			wg.Add(1)
			go func(b types.Build) {
				defer wg.Done()
				input, err := cb.ConvertBuildConfigToStartBuildInput(b)
				if err != nil {
					errChan <- fmt.Errorf("failed to convert build config for %s: %w", b.ProjectName, err)
					return
				}
				id, err := cb.RunCodeBuild(client, input)
				if err != nil {
					errChan <- fmt.Errorf("failed to start build for %s: %w", b.ProjectName, err)
				} else {
					idsChan <- id
				}
			}(build)
		}

		wg.Wait()
		close(idsChan)
		close(errChan)

		// Collect results
		ids := []string{}
		runErrors := []error{}
		for id := range idsChan {
			ids = append(ids, id)
		}
		for err := range errChan {
			log.Println(err) // Log each run error immediately
			runErrors = append(runErrors, err)
		}

		// Exit if there were errors starting builds
		if len(runErrors) > 0 {
			log.Printf("%d build(s) failed to start.", len(runErrors))
			os.Exit(1)
		}

		// Early return if --no-wait option set or no builds were successfully started
		if nowait || len(ids) == 0 {
			if len(ids) == 0 {
				log.Println("No builds were started successfully.")
			}
			return
		}

		// Check build status
		failed := false
		failed, err = cb.WaitAndCheckBuildStatus(client, ids, pollsec)
		if err != nil {
			log.Fatal(err)
		}

		// Exit with non-zero code if any build failed (either starting or during run)
		if failed { // WaitAndCheckBuildStatus indicates a failure during run
			os.Exit(2)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&nowait, "no-wait", false, "specify if you don't need to follow builds status")
	runCmd.Flags().IntVar(&pollsec, "polling-span", 60, "polling span in second for builds status check")
	runCmd.Flags().StringSliceVar(&targets, "targets", []string{}, "Specify target group(s) to run (only available for map format config)")
}
