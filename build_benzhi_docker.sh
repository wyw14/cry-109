#!/usr/bin/env sh
set -eu
docker build --platform linux/amd64 -f benzhi.Dockerfile -t portcrane:local .
