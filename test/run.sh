#!/bin/bash

set -e -o pipefail

DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"


# Versions under test are listed in k8s-versions.json, which is also used to
# build the CI matrix. The node images are tightly coupled to the Kind version
# and they must be updated together, using the images listed in the release
# notes for that Kind version. K8S_VERSION defaults to the first entry.
VERSIONS_FILE="${DIR}/k8s-versions.json"
KIND_VERSION=$(jq -r '.kind' "${VERSIONS_FILE}")
K8S_VERSION=${K8S_VERSION:-$(jq -r '.kubernetes[0].version' "${VERSIONS_FILE}")}
KIND_NODE_IMAGE=$(jq -r --arg v "$K8S_VERSION" '.kubernetes[] | select(.version == $v) | .nodeImage' "${VERSIONS_FILE}")
if [ -z "$KIND_NODE_IMAGE" ]; then
    echo "no kind node available for Kind $KIND_VERSION and Kubernetes $K8S_VERSION in ${VERSIONS_FILE}" 1>&2
    exit 1
fi

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
    arm64|aarch64) ARCH=arm64 ;;
    *)
        echo "unsupported ARCH: ${ARCH}" 1>&2
        exit 1
        ;;
esac

SUCCESS=
cleanup() {
    if [ -z "$SUCCESS" ]; then
        echo "================================================="
        echo "LOGS: SPIRE Server"
        echo "================================================="
        "${KUBECTL}" logs -nspire-system deployment/spire-server --all-containers=true || true

        echo "================================================="
        echo "LOGS: SPIRE Agent"
        echo "================================================="
        "${KUBECTL}" logs -nspire-system daemonset/spire-agent --all-containers=true || true

        echo "================================================="
        echo "LOGS: SPIFFE CSI Driver"
        echo "================================================="
        "${KUBECTL}" logs -nspire-system daemonset/spiffe-csi-driver --all-containers=true || true

        echo "================================================="
        echo "LOGS: Test Workload 1"
        echo "================================================="
        "${KUBECTL}" logs deployment/test-workload-1 --all-containers=true || true

        echo "================================================="
        echo "LOGS: Test Workload 2"
        echo "================================================="
        "${KUBECTL}" logs deployment/test-workload-2 --all-containers=true || true

        echo "================================================="
        echo "LOGS: SPIFFE CSI Driver (recursive bind)"
        echo "================================================="
        "${KUBECTL}" logs -nspire-system daemonset/spiffe-csi-driver-recursive --all-containers=true || true

        echo "================================================="
        echo "MOUNTS: recursive bind, as the node sees them"
        echo "================================================="
        on-node grep -E "spiffe-csi-test|recursive" /proc/self/mountinfo || true

    fi

    [ -n "$SKIPCLEANUP" ] || delete-cluster
    [ -n "$SKIPCLEANUP" ] || rm -rf "${TMPDIR}"

    if [ -z "$SUCCESS" ]; then
        echo "================================================="
        echo "!!! FAILED !!!!"
        echo "================================================="
    fi
}

build-workload() {
    (cd "${DIR}"/workload; docker build -t spiffe-csi-test-workload:test .)
}

download-kubectl() {
    local _url="https://dl.k8s.io/release/${K8S_VERSION}/bin/${OS}/${ARCH}/kubectl"
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
    "${KIND}" create cluster --image="$KIND_NODE_IMAGE" --config "${DIR}/config/cluster.yaml"
    echo "Cluster created."
    "${KUBECTL}" version
}

delete-cluster() {
    echo "Deleting cluster..."
    "${KIND}" delete cluster
    echo "Cluster deleted."
}

load-images() {
    org_name=$(echo "$GITHUB_REPOSITORY" | tr '/' "\n" | head -1 | tr -d "\n")
    org_name="${org_name:-spiffe}" # default to spiffe in case ran on local
    registry=ghcr.io/${org_name}

    echo "Loading images..."
    "${KIND}" load docker-image \
        "${registry}/spiffe-csi-driver:devel" \
        spiffe-csi-test-workload:test
    echo "Images loaded."
}

deploy-spire() {
    "${KUBECTL}" apply -k "${DIR}"/config/spire
    echo "Waiting for SPIRE server rollout..."
    "${KUBECTL}" rollout status -w --timeout=1m -nspire-system deployment/spire-server
    echo "Waiting for SPIRE bundle to be published to ConfigMap..."
    until "${KUBECTL}" get configmap spire-bundle -nspire-system -o jsonpath='{.data.bundle\.crt}' 2>/dev/null | grep -q .; do
        sleep 1
    done
    echo "SPIRE bundle is ready."
    echo "Waiting for SPIRE agent rollout..."
    "${KUBECTL}" rollout status -w --timeout=1m -nspire-system daemonset/spire-agent
    echo "Waiting for SPIFFE CSI Driver rollout..."
    "${KUBECTL}" rollout status -w --timeout=1m -nspire-system daemonset/spiffe-csi-driver
}

deploy-workloads() {
    "${KUBECTL}" apply -k "${DIR}"/config/workloads
    echo "Waiting for test workload 1 rollout..."
    "${KUBECTL}" rollout status -w --timeout=1m deployment/test-workload-1
    echo "Waiting for test workload 2 rollout..."
    "${KUBECTL}" rollout status -w --timeout=1m deployment/test-workload-2
}

register-workloads() {
    "${KUBECTL}" exec \
        -nspire-system \
        deployment/spire-server -- \
        /opt/spire/bin/spire-server entry create \
            -node \
            -spiffeID spiffe://test/node \
            -selector k8s_psat:cluster:test

    "${KUBECTL}" exec \
        -nspire-system \
        deployment/spire-server -- \
        /opt/spire/bin/spire-server entry create \
            -parentID spiffe://test/node \
            -spiffeID spiffe://test/workload \
            -selector k8s:ns:default
}

check-workload-status() {
    local _which="$1"
    local _numchecks=30 
    local _interval=1
    local _status
    echo -n "Checking Workload API update status on $_which."
    for ((i=1;i<=_numchecks;i++)); do
        echo -n "."
        _status=$("${KUBECTL}" exec "deployment/$_which" -- /bin/cat status | tr -d '[:space:]')
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

# The source directory the recursive driver publishes, on the node.
RECURSIVE_SOURCE=/run/spiffe-csi-test/recursive

# Runs a command on the node, in its own mount namespace. Kind nodes are
# containers on the machine running the test.
on-node() {
    docker exec "$("${KIND}" get nodes | head -n1)" "$@"
}

deploy-recursive() {
    "${KUBECTL}" apply -k "${DIR}"/config/recursive
    echo "Waiting for SPIFFE CSI Driver (recursive bind) rollout..."
    "${KUBECTL}" rollout status -w --timeout=1m -nspire-system daemonset/spiffe-csi-driver-recursive
    echo "Waiting for recursive workload 1 rollout..."
    "${KUBECTL}" rollout status -w --timeout=1m deployment/recursive-workload-1
    echo "Waiting for recursive workload 2 rollout..."
    "${KUBECTL}" rollout status -w --timeout=1m deployment/recursive-workload-2
}

# Waits for a workload to read the marker from the filesystem mounted beneath
# its volume. Propagation is synchronous, so this only waits out kubectl.
check-workload-sees-mount() {
    local _which="$1"
    local _numchecks=30
    local _interval=1
    local _got
    echo -n "Checking $_which sees the mount beneath its volume."
    for ((i=1;i<=_numchecks;i++)); do
        echo -n "."
        _got=$("${KUBECTL}" exec "deployment/$_which" -- /bin/cat /recursive/sub/marker 2>/dev/null | tr -d '[:space:]' || true)
        if [ "${_got}" == "propagated" ]; then
            echo "ok."
            return 0
        fi
        sleep "${_interval}"
    done
    echo "oh no!"
    echo "$_which does not see the mount beneath its volume: marker=${_got}" 1>&2
    return 1
}

check-node-has-mount() {
    if ! on-node grep -q " ${RECURSIVE_SOURCE}/sub " /proc/self/mountinfo; then
        echo "The node no longer has the mount at ${RECURSIVE_SOURCE}/sub" 1>&2
        return 1
    fi
    if [ "$(on-node cat "${RECURSIVE_SOURCE}/sub/marker" | tr -d '[:space:]')" != "propagated" ]; then
        echo "The node's mount at ${RECURSIVE_SOURCE}/sub no longer has the marker" 1>&2
        return 1
    fi
    echo "The node still has the mount."
}

# A filesystem mounted beneath the socket directory after two workloads have
# started has to reach both, and tearing one workload down has to leave it for
# the other, for the node, and for the driver's own view, which the replacement
# workload is published from. A recursive bind whose copies stay peers of the
# source fails here: detaching one workload's volume unmounts the filesystem
# for everyone.
check-recursive-bind() {
    echo "Checking the recursive bind..."

    # Nothing is mounted yet. If a workload already sees the marker, the
    # ordering under test is not happening.
    for _which in recursive-workload-1 recursive-workload-2; do
        if "${KUBECTL}" exec "deployment/$_which" -- /bin/sh -c 'test -e /recursive/sub/marker'; then
            echo "$_which sees the marker before anything was mounted" 1>&2
            return 1
        fi
    done

    on-node mkdir -p "${RECURSIVE_SOURCE}/sub"
    on-node mount -t tmpfs recursive-test "${RECURSIVE_SOURCE}/sub"
    on-node /bin/sh -c "echo propagated > ${RECURSIVE_SOURCE}/sub/marker"

    check-workload-sees-mount recursive-workload-1
    check-workload-sees-mount recursive-workload-2

    echo "Deleting the recursive workload 1 pod..."
    "${KUBECTL}" delete pod -l app=recursive-workload-1 --wait --timeout=1m

    check-workload-sees-mount recursive-workload-2
    check-node-has-mount

    echo "Waiting for the replacement recursive workload 1..."
    "${KUBECTL}" rollout status -w --timeout=1m deployment/recursive-workload-1
    check-workload-sees-mount recursive-workload-1
}

echo "Starting integration test..."

build-workload
download-kind
download-kubectl
create-cluster
load-images
deploy-spire
register-workloads
deploy-workloads
check-workload-status "test-workload-1"
check-workload-status "test-workload-2"
deploy-recursive
check-recursive-bind
SUCCESS=1
echo "Done."
