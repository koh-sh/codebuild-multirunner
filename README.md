# codebuild-multirunner

## Overview

This is a simple CLI tool to "Start build with overrides" multiple AWS CodeBuild Projects at once.

![Img](misc/ss.png)

## Installation

You can install with Homebrew.

```bash
brew install koh-sh/tap/codebuild-multirunner
```

Or download prebuild binary from [Releases](https://github.com/koh-sh/codebuild-multirunner/releases)

## Usage

```bash
This is a simple CLI tool to "Start build with overrides" multiple AWS CodeBuild Projects at once.

This command will read YAML based config file and run multiple CodeBuild projects with oneliner.

Usage:
  codebuild-multirunner [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  dump        dump config for running CodeBuild projects
  help        Help about any command
  log         Print CodeBuild log for a single build with a provided id.
  retry       retry CodeBuild build with a provided id
  run         run CodeBuild projects based on YAML

Flags:
      --config string   file path for config file. (default "./config.yaml")
  -h, --help            help for codebuild-multirunner
  -v, --version         version for codebuild-multirunner

Use "codebuild-multirunner [command] --help" for more information about a command.
```

## QuickStart

### Create Config file

Create YAML based config file.
change "testproject" to your CodeBuild Project name.

```bash
% cat config.yaml
builds:
  - projectName: testproject
```

### Run

Then execute command with "run" subcommand, so your CodeBuild project will be running.

```bash
codebuild-multirunner run
```

If you specify multiple projects, all projects will be running at once.

```bash
% cat config.yaml
builds:
  - projectName: testproject
  - projectName: testproject2
  - projectName: testproject3
```

You can "Start build with overrides" by specifying parameters.

```bash
% cat config.yaml
builds:
  - projectName: testproject
  - projectName: testproject2
    environmentVariablesOverride:
    - name: TEST_VAR
      value: FOOBAR
      type: PLAINTEXT
  - projectName: testproject3
```

Also environment variables are substituted for execution.

```bash
builds:
- projectName: testproject
- projectName: testproject2
  environmentVariablesOverride:
    - name: TEST_VAR
      value: FOOBAR
      type: PLAINTEXT
- projectName: testproject3
  sourceVersion: ${BRANCH_NAME} # it will read environment variable
```

You can check the config by "dump" subcommand.

```bash
% export BRANCH_NAME=feature/new_function
% codebuild-multirunner dump
builds:
    - projectName: testproject
    - environmentVariablesOverride:
        - name: TEST_VAR
          type: PLAINTEXT
          value: FOOBAR
      projectName: testproject2
    - projectName: testproject3
      sourceVersion: feature/new_function

%
```

Refer to [sample config file](config.yaml)

### Get build log

And `log` is useful to get detail of a build.

CloudWatch Logs need to be ENABLED for the build.

```bash
% codebuild-multirunner log --id testproject:33719fff-7ee7-4828-9c6a-ec814226e3fc
[Container] 2023/08/11 02:53:03 Waiting for agent ping
[Container] 2023/08/11 02:53:04 Waiting for DOWNLOAD_SOURCE
[Container] 2023/08/11 02:53:05 Phase is DOWNLOAD_SOURCE
[Container] 2023/08/11 02:53:05 CODEBUILD_SRC_DIR=/codebuild/output/src3794654258/src
...
(omitted)
...
[Container] 2023/08/11 02:53:15 Phase context status code:  Message:
[Container] 2023/08/11 02:53:15 Entering phase POST_BUILD
[Container] 2023/08/11 02:53:15 Phase complete: POST_BUILD State: SUCCEEDED
[Container] 2023/08/11 02:53:15 Phase context status code:  Message:
%
```

### Retry past builds

You can retry a past build.

```bash
% codebuild-multirunner retry --id testproject:8948df1b-1352-4f87-bc68-318a37a7949b
2023/08/19 14:52:28 testproject:dd3bd981-59ab-4c78-a0f2-22c75545ffc7 [STARTED]
2023/08/19 14:53:28 testproject:dd3bd981-59ab-4c78-a0f2-22c75545ffc7 [SUCCEEDED]
```
