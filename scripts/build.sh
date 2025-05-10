#!/usr/bin/env sh
set -eu

scriptDir=$(dirname "$(readlink -f "$0")")
rootDir=$scriptDir/..

appVersion="$(cat "$rootDir/VERSION")"
commitHash="$(git rev-parse HEAD)"
#commitBranch="$(git branch | grep "^\*" | sed 's/^..//')"
buildDate="$(date -u +'%Y%m%dT%H%M%SZ')"

go build -C "$rootDir/src" -v -ldflags "-s -w -X main.appVersion=$appVersion -X main.commitHash=$commitHash -X main.buildDate=$buildDate"
