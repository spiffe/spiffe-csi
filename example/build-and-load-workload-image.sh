#!/bin/bash

set -e -o pipefail

DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

echo "Building the example workload image..."
cd "$DIR"/workload
KO_DOCKER_REPO=ko.local/spiffe-csi-driver-example-workload ko build --bare -t example .

echo "Loading example workload image into Kind..."
kind load docker-image ko.local/spiffe-csi-driver-example-workload:example

echo "Done."
