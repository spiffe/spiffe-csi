# SPIFFE CSI Driver
**WARNING:** This project is in the "[Development](https://github.com/spiffe/spiffe/blob/main/MATURITY.md#development)" phase of the [SPIFFE Project Maturity Phases](https://github.com/spiffe/spiffe/blob/main/MATURITY.md).

A [Container Storage Interface](https://github.com/container-storage-interface/spec/blob/master/spec.md) driver for Kubernetes that facilitates injection of the SPIFFE Workload API.

The SPIFFE Workload API is nominally served over a Unix domain socket. Some SPIFFE implementations (e.g. SPIRE) rely on daemonsets to run one workload API server instance per host. In these cases, it is necessary to inject the workload API socket into each pod. The primary motivation for using a CSI driver for this purpose is to avoid the use of hostPath, which is associated with security weaknesses and is commonly disallowed by policy.

This driver provides pods with an ephemeral inline volume. SPIFFE implementations can serve their workload API socket in a central location, and the driver will bind mount this location into workload pods as directed.
