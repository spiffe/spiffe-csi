#!/bin/bash

set -e -o pipefail

DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

echo "Building the example workload image..."
docker build "$DIR"/workload -t spiffe-csi-driver-example-workload:example

echo "Loading example workload image into Kind..."
kind load docker-image spiffe-csi-driver-example-workload:example

echo "Done."
