#!/usr/bin/env sh
set -eu

scriptDir=$(dirname "$(readlink -f "$0")")

platform=${1:-}
suffix=${2:-}
releaseTag=${3:-}

if [ -z "$releaseTag" ]; then
	echo "Usage: $0 PLATFORM SUFFIX RELEASE_TAG"
	echo
	echo Example: "$0" linux/arm/v7 arm32v7 latest
	exit 1
fi

pluginId="anthochamp/vaultfs-${suffix}"

"$scriptDir/create.sh" "$platform" "$pluginId" "$releaseTag"

docker --context default plugin push "$pluginId:$releaseTag"
docker --context default plugin rm "$pluginId:$releaseTag"
