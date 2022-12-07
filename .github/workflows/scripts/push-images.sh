#!/usr/bin/env bash
# shellcheck shell=bash
##
## USAGE: __PROG__
##
## "__PROG__" publishes images to a registry.
##
## Usage example(s):
##   ./__PROG__ 1.5.2
##   ./__PROG__ v1.5.2
##   ./__PROG__ refs/tags/v1.5.2
##
## Commands
## - ./__PROG__ <version>    pushes images to the registry using given version.

set -e

function usage {
  grep '^##' "$0" | sed -e 's/^##//' -e "s/__PROG__/$me/" >&2
}

me=$(basename "$0")

version="$1"
# remove the git tag prefix
# Push the images using the version tag (without the "v" prefix).
# Also strips the refs/tags part if the GITHUB_REF variable is used.
version="${version#refs/tags/v}"
version="${version#v}"

if [ -z "${version}" ]; then
    usage
    echo "version not provided!" 1>&2
    exit 1
fi

echo "Pushing image tagged as ${version}..."

LOCALIMG=ghcr.io/spiffe/spiffe-csi-driver:devel
REMOTEIMG=ghcr.io/spiffe/spiffe-csi-driver:"${version}"

echo "Executing: docker tag $LOCALIMG $REMOTEIMG"
docker tag "$LOCALIMG" "$REMOTEIMG"
echo "Executing: docker push $REMOTEIMG"
docker push "$REMOTEIMG"
