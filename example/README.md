# SPIFFE CSI Driver Example

This example demonstrates how to deploy the SPIFFE CSI Driver into a Kubernetes
cluster and how to consume the Workload API Unix Domain Socket it provides
from a SPIFFE-aware workload.

## Prerequisites

- [Kind](https://kind.sigs.k8s.io/)
- [Kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)

## Steps

1. Start a Kubernetes cluster via [Kind](https://kind.sigs.k8s.io/):

    ```
    $ kind create cluster
    ```

1. Build the example workload image and load it into Kind:

    ```
    $ ./build-and-load-workload-image.sh
    ```

1. Deploy SPIRE and the SPIFFE CSI Driver (which resides in the same DaemonSet as the SPIRE Agent):

    ```
    $ ./deploy-spire-and-csi-driver.sh
    ```

1. Register the example workload with SPIRE Server:

    ```
    $ ./register-workload.sh
    ```

1. Deploy the workload:

    ```
    $ kubectl apply -f config/workload.yaml
    ```

1. Check the workload logs to see the update received over the Workload API:

    ```
    $ kubectl logs pod/example-workload
    ```

    You should see something like:

    ```
    2021/11/23 18:46:33 Update:
    2021/11/23 18:46:33   SVIDs:
    2021/11/23 18:46:33     spiffe://example.org/workload
    2021/11/23 18:46:33   Bundles:
    2021/11/23 18:46:33     example.org (1 authorities)
    ```

1. Delete the Kubernetes cluster:

    ```
    $ kind delete cluster
    ```
