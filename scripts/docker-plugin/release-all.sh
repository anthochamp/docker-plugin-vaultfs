#!/usr/bin/env sh
set -eu

scriptDir=$(dirname "$(readlink -f "$0")")

releaseTag=${1:-}

if [ -z "$releaseTag" ]; then
	echo "Usage: $0 RELEASE_TAG"
	echo
	echo Example: "$0" latest
	exit 1
fi

"$scriptDir/release.sh" linux/amd64 amd64 "$releaseTag"
"$scriptDir/release.sh" linux/arm/v6 arm32v6 "$releaseTag"
"$scriptDir/release.sh" linux/arm/v7 arm32v7 "$releaseTag"
"$scriptDir/release.sh" linux/arm64/v8 arm64v8 "$releaseTag"
"$scriptDir/release.sh" linux/386 i386 "$releaseTag"
"$scriptDir/release.sh" linux/ppc64le ppc64le "$releaseTag"
"$scriptDir/release.sh" linux/riscv64 riscv64 "$releaseTag"
"$scriptDir/release.sh" linux/s390x s390x "$releaseTag"
