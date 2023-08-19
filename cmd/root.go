package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// options
var Configfile string

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "codebuild-multirunner",
	Short: "This is a simple CLI tool to \"Start build with overrides\" multiple AWS CodeBuild Projects at once.",
	Long: `This is a simple CLI tool to "Start build with overrides" multiple AWS CodeBuild Projects at once.

This command will read YAML based config file and run multiple CodeBuild projects with oneliner.
`,
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&Configfile, "config", "./config.yaml", "file path for config file.")
}

// set version from goreleaser variables
func SetVersionInfo(version, commit, date string) {
	RootCmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)
}

//
// types and functions shared within subcommands
//

// interface for AWS CodeBuild API
type CodeBuildAPI interface {
	BatchGetBuilds(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error)
	StartBuild(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error)
	RetryBuild(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error)
}

// interface for AWS CloudWatch Logs API
type CWLGetLogEventsAPI interface {
	GetLogEvents(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error)
}

// return CodeBuild api client
func NewCodeBuildAPI() (CodeBuildAPI, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	return codebuild.NewFromConfig(cfg), nil
}

// return CloudWatchLogs api client
func NewCloudWatchLogsAPI() (CWLGetLogEventsAPI, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	return cloudwatchlogs.NewFromConfig(cfg), nil
}

// read yaml config file for builds definition
func ReadConfigFile(filepath string) BuildConfig {
	bc := BuildConfig{}
	b, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	expanded := os.ExpandEnv(string(b))
	err = yaml.Unmarshal([]byte(expanded), &bc)
	if err != nil {
		log.Fatal(err)
	}
	return bc
}

//
// types for CodeBuild parameter override. defined to use yaml tags
// https://docs.aws.amazon.com/codebuild/latest/APIReference/API_StartBuild.html
//

// list of Builds
type BuildConfig struct {
	Builds []Build `yaml:"builds"`
}

type ArtifactsOverride struct {
	ArtifactIdentifier   string `yaml:"artifactIdentifier,omitempty"`
	BucketOwnerAccess    string `yaml:"bucketOwnerAccess,omitempty"`
	EncryptionDisabled   bool   `yaml:"encryptionDisabled,omitempty"`
	Location             string `yaml:"location,omitempty"`
	Name                 string `yaml:"name,omitempty"`
	NamespaceType        string `yaml:"namespaceType,omitempty"`
	OverrideArtifactName bool   `yaml:"overrideArtifactName,omitempty"`
	Packaging            string `yaml:"packaging,omitempty"`
	Path                 string `yaml:"path,omitempty"`
	Type                 string `yaml:"type,omitempty"`
}
type BuildStatusConfigOverride struct {
	Context   string `yaml:"context,omitempty"`
	TargetURL string `yaml:"targetUrl,omitempty"`
}
type CacheOverride struct {
	Location string   `yaml:"location,omitempty"`
	Modes    []string `yaml:"modes,omitempty"`
	Type     string   `yaml:"type,omitempty"`
}
type EnvironmentVariablesOverride struct {
	Name  string `yaml:"name,omitempty"`
	Type  string `yaml:"type,omitempty"`
	Value string `yaml:"value,omitempty"`
}
type GitSubmodulesConfigOverride struct {
	FetchSubmodules bool `yaml:"fetchSubmodules,omitempty"`
}
type CloudWatchLogs struct {
	GroupName  string `yaml:"groupName,omitempty"`
	Status     string `yaml:"status,omitempty"`
	StreamName string `yaml:"streamName,omitempty"`
}
type S3Logs struct {
	BucketOwnerAccess  string `yaml:"bucketOwnerAccess,omitempty"`
	EncryptionDisabled bool   `yaml:"encryptionDisabled,omitempty"`
	Location           string `yaml:"location,omitempty"`
	Status             string `yaml:"status,omitempty"`
}
type LogsConfigOverride struct {
	CloudWatchLogs CloudWatchLogs `yaml:"cloudWatchLogs,omitempty"`
	S3Logs         S3Logs         `yaml:"s3Logs,omitempty"`
}
type RegistryCredentialOverride struct {
	Credential         string `yaml:"credential,omitempty"`
	CredentialProvider string `yaml:"credentialProvider,omitempty"`
}
type SecondaryArtifactsOverride struct {
	ArtifactIdentifier   string `yaml:"artifactIdentifier,omitempty"`
	BucketOwnerAccess    string `yaml:"bucketOwnerAccess,omitempty"`
	EncryptionDisabled   bool   `yaml:"encryptionDisabled,omitempty"`
	Location             string `yaml:"location,omitempty"`
	Name                 string `yaml:"name,omitempty"`
	NamespaceType        string `yaml:"namespaceType,omitempty"`
	OverrideArtifactName bool   `yaml:"overrideArtifactName,omitempty"`
	Packaging            string `yaml:"packaging,omitempty"`
	Path                 string `yaml:"path,omitempty"`
	Type                 string `yaml:"type,omitempty"`
}
type Auth struct {
	Resource string `yaml:"resource,omitempty"`
	Type     string `yaml:"type,omitempty"`
}
type BuildStatusConfig struct {
	Context   string `yaml:"context,omitempty"`
	TargetURL string `yaml:"targetUrl,omitempty"`
}
type GitSubmodulesConfig struct {
	FetchSubmodules bool `yaml:"fetchSubmodules,omitempty"`
}
type SecondarySourcesOverride struct {
	Auth                Auth                `yaml:"auth,omitempty"`
	Buildspec           string              `yaml:"buildspec,omitempty"`
	BuildStatusConfig   BuildStatusConfig   `yaml:"buildStatusConfig,omitempty"`
	GitCloneDepth       int                 `yaml:"gitCloneDepth,omitempty"`
	GitSubmodulesConfig GitSubmodulesConfig `yaml:"gitSubmodulesConfig,omitempty"`
	InsecureSsl         bool                `yaml:"insecureSsl,omitempty"`
	Location            string              `yaml:"location,omitempty"`
	ReportBuildStatus   bool                `yaml:"reportBuildStatus,omitempty"`
	SourceIdentifier    string              `yaml:"sourceIdentifier,omitempty"`
	Type                string              `yaml:"type,omitempty"`
}
type SecondarySourcesVersionOverride struct {
	SourceIdentifier string `yaml:"sourceIdentifier,omitempty"`
	SourceVersion    string `yaml:"sourceVersion,omitempty"`
}
type SourceAuthOverride struct {
	Resource string `yaml:"resource,omitempty"`
	Type     string `yaml:"type,omitempty"`
}
type Build struct {
	ArtifactsOverride                ArtifactsOverride                 `yaml:"artifactsOverride,omitempty"`
	BuildspecOverride                string                            `yaml:"buildspecOverride,omitempty"`
	BuildStatusConfigOverride        BuildStatusConfigOverride         `yaml:"buildStatusConfigOverride,omitempty"`
	CacheOverride                    CacheOverride                     `yaml:"cacheOverride,omitempty"`
	CertificateOverride              string                            `yaml:"certificateOverride,omitempty"`
	ComputeTypeOverride              string                            `yaml:"computeTypeOverride,omitempty"`
	DebugSessionEnabled              bool                              `yaml:"debugSessionEnabled,omitempty"`
	EncryptionKeyOverride            string                            `yaml:"encryptionKeyOverride,omitempty"`
	EnvironmentTypeOverride          string                            `yaml:"environmentTypeOverride,omitempty"`
	EnvironmentVariablesOverride     []EnvironmentVariablesOverride    `yaml:"environmentVariablesOverride,omitempty"`
	GitCloneDepthOverride            int                               `yaml:"gitCloneDepthOverride,omitempty"`
	GitSubmodulesConfigOverride      GitSubmodulesConfigOverride       `yaml:"gitSubmodulesConfigOverride,omitempty"`
	IdempotencyToken                 string                            `yaml:"idempotencyToken,omitempty"`
	ImageOverride                    string                            `yaml:"imageOverride,omitempty"`
	ImagePullCredentialsTypeOverride string                            `yaml:"imagePullCredentialsTypeOverride,omitempty"`
	InsecureSslOverride              bool                              `yaml:"insecureSslOverride,omitempty"`
	LogsConfigOverride               LogsConfigOverride                `yaml:"logsConfigOverride,omitempty"`
	PrivilegedModeOverride           bool                              `yaml:"privilegedModeOverride,omitempty"`
	ProjectName                      string                            `yaml:"projectName"`
	QueuedTimeoutInMinutesOverride   int                               `yaml:"queuedTimeoutInMinutesOverride,omitempty"`
	RegistryCredentialOverride       RegistryCredentialOverride        `yaml:"registryCredentialOverride,omitempty"`
	ReportBuildStatusOverride        bool                              `yaml:"reportBuildStatusOverride,omitempty"`
	SecondaryArtifactsOverride       []SecondaryArtifactsOverride      `yaml:"secondaryArtifactsOverride,omitempty"`
	SecondarySourcesOverride         []SecondarySourcesOverride        `yaml:"secondarySourcesOverride,omitempty"`
	SecondarySourcesVersionOverride  []SecondarySourcesVersionOverride `yaml:"secondarySourcesVersionOverride,omitempty"`
	ServiceRoleOverride              string                            `yaml:"serviceRoleOverride,omitempty"`
	SourceAuthOverride               SourceAuthOverride                `yaml:"sourceAuthOverride,omitempty"`
	SourceLocationOverride           string                            `yaml:"sourceLocationOverride,omitempty"`
	SourceTypeOverride               string                            `yaml:"sourceTypeOverride,omitempty"`
	SourceVersion                    string                            `yaml:"sourceVersion,omitempty"`
	TimeoutInMinutesOverride         int                               `yaml:"timeoutInMinutesOverride,omitempty"`
}

//
// some types and functions for AWS SDK Mock. used only for testing
//

// mock api for StartBuild
type StartBuildMockAPI func(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error)

func (m StartBuildMockAPI) StartBuild(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error) {
	return m(ctx, params, optFns...)
}

func (m StartBuildMockAPI) BatchGetBuilds(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
	return nil, nil
}

func (m StartBuildMockAPI) RetryBuild(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error) {
	return nil, nil
}

// mock api for BatchGetBuilds
type BatchGetBuildsMockAPI func(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error)

func (m BatchGetBuildsMockAPI) StartBuild(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error) {
	return nil, nil
}

func (m BatchGetBuildsMockAPI) BatchGetBuilds(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
	return m(ctx, params, optFns...)
}

func (m BatchGetBuildsMockAPI) RetryBuild(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error) {
	return nil, nil
}

// mock api for BatchGetBuilds
type RetryBuildMockAPI func(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error)

func (m RetryBuildMockAPI) StartBuild(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error) {
	return nil, nil
}

func (m RetryBuildMockAPI) BatchGetBuilds(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
	return nil, nil
}

func (m RetryBuildMockAPI) RetryBuild(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error) {
	return m(ctx, params, optFns...)
}

// mock api for GetLogEvents
type GetLogEventsMockAPI func(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error)

func (m GetLogEventsMockAPI) GetLogEvents(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error) {
	return m(ctx, params, optFns...)
}

// return mock function for StartBuild
func ReturnStartBuildMockAPI(build *types.Build, err error) func(t *testing.T) CodeBuildAPI {
	mock := func(t *testing.T) CodeBuildAPI {
		return StartBuildMockAPI(func(ctx context.Context, params *codebuild.StartBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.StartBuildOutput, error) {
			t.Helper()
			if params.ProjectName == nil {
				t.Fatal("ProjectName is necessary")
			}
			return &codebuild.StartBuildOutput{
				Build:          build,
				ResultMetadata: middleware.Metadata{},
			}, err
		})
	}
	return mock
}

// return mock function for BatchgetBuilds
func ReturnBatchGetBuildsMockAPI(builds []types.Build) func(t *testing.T) CodeBuildAPI {
	mock := func(t *testing.T) CodeBuildAPI {
		return BatchGetBuildsMockAPI(func(ctx context.Context, params *codebuild.BatchGetBuildsInput, optFns ...func(*codebuild.Options)) (*codebuild.BatchGetBuildsOutput, error) {
			t.Helper()
			if len(params.Ids) == 0 {
				t.Fatal("Ids must have at least one")
			}
			return &codebuild.BatchGetBuildsOutput{
				Builds:         builds,
				BuildsNotFound: []string{},
				ResultMetadata: middleware.Metadata{},
			}, nil
		})
	}
	return mock
}

// return mock function for GetLogEvents
func ReturnGetLogEventsMockAPI(events []cwltypes.OutputLogEvent) func(t *testing.T) CWLGetLogEventsAPI {
	mock := func(t *testing.T) CWLGetLogEventsAPI {
		return GetLogEventsMockAPI(func(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error) {
			t.Helper()
			if *params.LogGroupName == "" || *params.LogStreamName == "" {
				t.Fatal("you must supply a logGroupName and logStreamName")
			}
			return &cloudwatchlogs.GetLogEventsOutput{
				Events:            events,
				NextBackwardToken: nil,
				NextForwardToken:  nil,
			}, nil
		})
	}
	return mock
}

// return mock function for BatchgetBuilds
func ReturnRetryBuildMockAPI(build types.Build) func(t *testing.T) CodeBuildAPI {
	mock := func(t *testing.T) CodeBuildAPI {
		return RetryBuildMockAPI(func(ctx context.Context, params *codebuild.RetryBuildInput, optFns ...func(*codebuild.Options)) (*codebuild.RetryBuildOutput, error) {
			t.Helper()
			if *params.Id == "" {
				t.Fatal("Id must have at least one")
			}
			return &codebuild.RetryBuildOutput{
				Build:          &build,
				ResultMetadata: middleware.Metadata{},
			}, nil
		})
	}
	return mock
}
