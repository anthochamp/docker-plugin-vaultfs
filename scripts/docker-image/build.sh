#!/usr/bin/env sh
set -eu

scriptDir=$(dirname "$(readlink -f "$0")")
rootDir=$scriptDir/../..

platform=${1:-}
tag=${2:-}

if [ -z "$tag" ]; then
	echo "Usage: $0 <platform> <tag>"
	echo
	echo Example:
	echo "  $0 linux/amd64 vaultfs:latest"
	exit 1
fi

appVersion="$(cat "$rootDir/VERSION")"
commitHash="$(git rev-parse HEAD)"
commitBranch="$(git branch | grep "^\*" | sed 's/^..//')"
buildDate="$(date -u +'%Y%m%dT%H%M%SZ')"

instance=$(docker --context default buildx create --use)

docker --context default buildx build \
	--platform "$platform" --load \
	--build-arg appVersion="$appVersion" \
	--build-arg commitHash="$commitHash" \
	--build-arg commitBranch="$commitBranch" \
	--build-arg buildDate="$buildDate" \
	--tag "$tag" \
	--file "$rootDir/packages/docker-image/Dockerfile" \
	"$rootDir"

docker --context default buildx rm "$instance"
