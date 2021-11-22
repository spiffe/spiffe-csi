#!/bin/bash

set -e -o pipefail

DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# Versions under test
SPIFFE_CSI_DRIVER_VERSION=${SPIFFE_CSI_DRIVER_VERSION:-devel}
KIND_VERSION=${KIND_VERSION:-v0.11.1}
KUBECTL_VERSION=${KUBECTL_VERSION:-v1.22.4}

# Export the Kind cluster name so we don't have to specify it on every kind
# invocation
KIND_CLUSTER_NAME=spiffe-csi-test
export KIND_CLUSTER_NAME

# Create a temporary directory for everything to live in
TMPDIR=$(mktemp -d)
trap cleanup EXIT
KIND="${TMPDIR}"/kind
KUBECTL="${TMPDIR}"/kubectl

# Figure out the OS and ARCH we're running under
OS=$(uname -s)
case "${OS}" in
    Darwin) OS=darwin ;;
    Linux) OS=linux ;;
    *)
        echo "unsupported OS: ${OS}" 1>&2
        exit 1
        ;;
esac

ARCH=$(uname -m)
case "${ARCH}" in
    x86_64) ARCH=amd64 ;;
    aarch64) ARCH2=arm64 ;;
    *)
        echo "unsupported ARCH: ${ARCH}" 1>&2
        exit 1
        ;;
esac


cleanup() {
    delete-cluster
    rm -rf "${TMPDIR}"
}

build-workload() {
    (cd "${DIR}"/workload; docker build -t spiffe-csi-test-workload:test .)
}

download-kubectl() {
    local _url="https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/${OS}/${ARCH}/kubectl"
    echo "Downloading ${_url}..."
    curl -s -Lo "${KUBECTL}" "${_url}"
    chmod +x "${KUBECTL}"
}

download-kind() {
    local _url="https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-${OS}-${ARCH}"
    echo "Downloading ${_url}..."
    curl -s -Lo "${KIND}" "${_url}"
    chmod +x "${KIND}"
}

create-cluster() {
    echo "Creating cluster..."
    "${KIND}" create cluster
    echo "Cluster created."
}

delete-cluster() {
    echo "Deleting cluster..."
    "${KIND}" delete cluster
    echo "Cluster deleted."
}

load-images() {
    echo "Loading images..."
    "${KIND}" load docker-image \
        ghcr.io/spiffe/spiffe-csi-driver:"${SPIFFE_CSI_DRIVER_VERSION}" \
        spiffe-csi-test-workload:test
    echo "Images loaded."
}

apply-yaml() {
    "${KUBECTL}" apply -k "${DIR}"/config
    echo "Waiting for SPIRE server rollout..."
    "${KUBECTL}" rollout status -w --timeout=1m -nspire-system deployment/spire-server
    echo "Waiting for SPIRE agent rollout..."
    "${KUBECTL}" rollout status -w --timeout=1m -nspire-system daemonset/spire-agent
    echo "Waiting for test workload rollout..."
    "${KUBECTL}" rollout status -w --timeout=1m deployment/test-workload
}

register-workload() {
    "${KUBECTL}" exec -it \
        -nspire-system \
        deployment/spire-server -- \
        /opt/spire/bin/spire-server entry create \
            -node \
            -spiffeID spiffe://test/node \
            -selector k8s_psat:cluster:test

    "${KUBECTL}" exec -it \
        -nspire-system \
        deployment/spire-server -- \
        /opt/spire/bin/spire-server entry create \
            -parentID spiffe://test/node \
            -spiffeID spiffe://test/workload \
            -selector k8s:ns:default
}

check-status() {
    local _numchecks=30 
    local _interval=1
    local _status
    echo -n "Checking Workload API update status."
    for ((i=1;i<=_numchecks;i++)); do
        echo -n "."
        _status=$(kubectl exec -it deployment/test-workload -- /bin/cat status | tr -d '[:space:]')
        if [ "${_status}" == "updated" ]; then
            echo "ok."
            return 0
        fi
        sleep "${_interval}"
    done
    echo "oh no!"
    echo "Failed to receive Workload API update in time: status=$_status" 1>&2
    return 1
}

echo "Starting integration test..."

build-workload
download-kind
download-kubectl
create-cluster
load-images
apply-yaml
register-workload
check-status

echo "Done."
