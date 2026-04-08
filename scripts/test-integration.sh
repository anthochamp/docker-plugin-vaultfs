#!/usr/bin/env sh
set -eu

scriptDir=$(dirname "$(readlink -f "$0")")
rootDir=$scriptDir/..

composeFile=$rootDir/test/docker-compose.test.yml

cleanup() {
  docker compose -f "$composeFile" down --volumes --remove-orphans
}

trap cleanup EXIT INT TERM

docker compose -f "$composeFile" up --detach --wait

VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=myroot \
  go test -tags integration -v -count=1 "$rootDir/..."
