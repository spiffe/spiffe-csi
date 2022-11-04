# SPIFFE CSI Driver

[![Pre-Production Phase](https://github.com/spiffe/spiffe/blob/main/.img/maturity/pre-prod.svg)](https://github.com/spiffe/spiffe/blob/main/MATURITY.md#pre-production)

A [Container Storage
Interface](https://github.com/container-storage-interface/spec/blob/master/spec.md)
driver for Kubernetes that facilitates injection of the SPIFFE Workload API.

The SPIFFE Workload API is nominally served over a Unix domain socket. Some
SPIFFE implementations (e.g. [SPIRE](https://github.com/spiffe/spire)) rely on DaemonSets to run one Workload API
server instance per host. In these cases, it is necessary to inject the
Workload API socket into each pod. The primary motivation for using a CSI
driver for this purpose is to avoid the use of
[hostPath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath)
volumes in workload containers, which is commonly disallowed or limited by
policy due to inherent security concerns. Note that `hostPath` volumes are
still required for the CSI driver to interact with the
[Kubelet](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/)
(see [Limitations](#limitations)).

This driver mounts a directory containing a SPIFFE Workload API socket provided
by a SPIFFE implementation (e.g. SPIRE) as an ephemeral inline volume into
workload pods.

## How it Works

This component is generally deployed as a container in the DaemonSet that
provides the Workload API implementation (e.g SPIRE agent) and registered with
the kubelet using the official CSI Node Driver Registrar image. The SPIFFE CSI
Driver and the Workload API implementation share the directory hosting the
Workload API Unix Domain Socket using an `hostPath` volume. An `emptyDir`
volume cannot be used since the backing directory would be removed if the
SPIFFE CSI Driver pod is restarted, invalidating the mount into workload
containers.

When pods declare an ephemeral inline mount using this driver, the driver is
invoked to mount the volume. The driver does a read-only bind mount of the
directory containing the Workload API Unix Domain Socket into the container
at the requested target path.

Similarly, when the pod is destroyed, the driver is invoked and removes the
bind mount.

## Dependencies

CSI Ephemeral Inline Volumes require at least Kubernetes 1.15 (enabled via the
`CSIInlineVolume` feature gate) or 1.16 (enabled by default).

## Limitations

CSI drivers are registered as plugins and otherwise interact with the Kubelet,
which requires several `hostPath` volumes. As such, this driver cannot be used
in environments where `hostPath` volumes are forbidden.

## Example

An example deployment can be found [here](./example). 

## Troubleshooting

This component has a fairly simple design and function but some of the
following problems may manifest.

### Failure to Register with the Kubelet

This problem can be diagnosed by dumping the logs of the kubelet (if possible),
the driver registrar container, and the SPIFFE CSI driver container. Likely
suspects are a misconfiguratoin of the various volume mounts needed for
communication between the register, the SPIFFE CSI driver, and the kubelet.

### Failure to Mount the Socket Directory

This problem can be diagnosed by dumping the SPIFFE CSI driver logs.

### Failure to Terminate Pods when Driver is Unhealthy Or Removed

If the SPIFFE CSI Driver is removed (or is otherwise unhealthy), any pods that
contain a volume mounted by the driver will fail to fully terminate until
driver health is restored. The describe command (i.e. kubectl describe) will
show the failure to unmount the volume. Kubernetes will continue to retry to
unmount the volume via the CSI driver. Once the driver has been restored, the
unmounting will eventually succeed and the pod will be fully terminated.

### Broken Mount when the CSI Driver Pod is Restarted

Ensure that the Workload API socket directory is shared with the SPIFFE CSI
Driver via a `hostPath` volume. The directory backing `emptyDir` volumes are
tied to the pod instance and invalidated when the pod is restarted.

## Reporting a Vulnerability

Vulnerabilities can be reported by sending an email to security@spiffe.io. A
confirmation email will be sent to acknowledge the report within 72 hours. A
second acknowledgement will be sent within 7 days when the vulnerability has
been positively or negatively confirmed.
