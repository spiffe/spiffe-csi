#!/bin/bash

set -e -o pipefail

DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"


###################################################################
# Caution: 
#
# This script registers the example workload with SPIRE so that it can obtain
# an SVID from the SPIFFE Workload API. Specifically, it creates an entry that
# groups all agents in the cluster and then registers any workload running in
# the default namespace against that group, effectively granting any workload
# in the default namespace the specified SPIFFE ID.
#
# These registrations for demonstrative purposes only and DO NOT reflect
# best practices for registering workloads inside of Kubernetes.
#
###################################################################

echo "Registering node..."
kubectl exec -it \
    -nspire \
    deployment/spire-server -- \
    /opt/spire/bin/spire-server entry create \
        -node \
        -spiffeID spiffe://example.org/node \
        -selector k8s_psat:cluster:example-cluster

echo "Registering workload..."
kubectl exec -it \
    -nspire \
    deployment/spire-server -- \
    /opt/spire/bin/spire-server entry create \
        -parentID spiffe://example.org/node \
        -spiffeID spiffe://example.org/workload \
        -selector k8s:ns:default

echo "Done."
