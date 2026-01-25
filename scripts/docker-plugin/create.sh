#!/usr/bin/env sh
set -eu

scriptDir=$(dirname "$(readlink -f "$0")")
rootDir=$scriptDir/../..

platform=${1:-}
pluginId=${2:-}
buildTag=${3:-latest}

if [ -z "$pluginId" ]; then
	echo Usage: "$(basename "$0")" PLATFORM PLUGIN_ID '[BUILD_TAG=latest]'
	echo
	echo Example:
	echo "  $0 linux/amd64 vaultfs latest"
	exit 1
fi

buildDir=$(mktemp -d)
imageName=$(mktemp -u dockerpluginrootfs-XXXXXXXX | awk '{ print tolower($0) }')

"$scriptDir/../docker-image/build.sh" "$platform" "$imageName"

containerId=$(docker --context default create "$imageName" true)

mkdir -p "$buildDir/rootfs"

docker --context default export "$containerId" | sudo tar --same-owner -p -x -C "$buildDir/rootfs"

docker --context default rm -vf "$containerId"
docker --context default rmi "$imageName"

cp "$rootDir/packages/docker-plugin/config.json" "$buildDir"

docker --context default plugin rm "$pluginId:$buildTag" >/dev/null 2>&1 || true
sudo docker --context default plugin create "$pluginId:$buildTag" "$buildDir"

sudo rm -r "$buildDir"
