package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// list of Builds
type BuildConfig struct {
	Builds []Build `yaml:"builds"`
}

// types for CodeBuild parameter override
// defined to use yaml tags
// https://docs.aws.amazon.com/codebuild/latest/APIReference/API_StartBuild.html
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

var configfile string

// read yaml config file for builds definition
func readConfigFile(filepath string) BuildConfig {
	bc := BuildConfig{}
	b, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	expanded := os.ExpandEnv(string(b))
	err = yaml.Unmarshal([]byte(expanded), &bc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return bc
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "codebuild-multirunner",
	Version: "v0.1.0",
	Short:   "This is a simple CLI tool to \"Start build with overrides\" multiple AWS CodeBuild Projects at once.",
	Long: `This is a simple CLI tool to "Start build with overrides" multiple AWS CodeBuild Projects at once.

This command will read YAML based config file and run multiple CodeBuild projects with oneliner.
`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configfile, "config", "./config.yaml", "file path for config file.")
}
