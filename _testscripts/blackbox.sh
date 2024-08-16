#!/usr/bin/env bash

#
# testing script for black box testing with real CodeBuild projects
#

set -eu

# test title colored with green
function title() {
    echo -e "\e[32m \n##### ${1} #####\n \e[m"
}

# fail message
function fail() {
    echo -e "\e[31m \n ${1} \n \e[m"
    exit 1
}

# go build
title "go build"
go build
ls -la ./codebuild-multirunner

# docker build
title "docker build"
docker build . -t codebuild-multirunner:latest
docker run -it --rm codebuild-multirunner:latest -v

# show help
title "help"
./codebuild-multirunner --help

# dump
title "dump"
./codebuild-multirunner dump --config "$(cd $(dirname $0);pwd)/codebuild-multirunner.yaml"

# run
title "run"
if ./codebuild-multirunner run --config "$(cd $(dirname $0);pwd)/codebuild-multirunner.yaml" --polling-span 5; then
    fail "return value should not be 0"
fi

# log
title "log"
LATEST=$(aws codebuild list-builds-for-project --project-name testproject | jq -r '.ids[]' | head -n 1)
./codebuild-multirunner log --id "$LATEST"

# retry
title "retry"
./codebuild-multirunner retry --id "$LATEST" --polling-span 10
