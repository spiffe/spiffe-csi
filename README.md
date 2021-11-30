# SPIFFE CSI Driver

[![Development Phase](https://img.shields.io/badge/SPIFFE-Dev-orange.svg?logoWidth=18&logo=data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHJvbGU9ImltZyIgdmlld0JveD0iMC4xMSAxLjg2IDM1OC4yOCAzNTguMjgiPjxzdHlsZT5zdmcge2VuYWJsZS1iYWNrZ3JvdW5kOm5ldyAwIDAgMzYwIDM2MH08L3N0eWxlPjxzdHlsZT4uc3QyLC5zdDN7ZmlsbC1ydWxlOmV2ZW5vZGQ7Y2xpcC1ydWxlOmV2ZW5vZGQ7ZmlsbDojYmNkOTE4fS5zdDN7ZmlsbDojMDRiZGQ5fTwvc3R5bGU+PGcgaWQ9IkxPR08iPjxwYXRoIGQ9Ik0xMi4xIDguOWgyOC4zYzIuNyAwIDUgMi4yIDUgNXYyOC4zYzAgMi43LTIuMiA1LTUgNUgxMi4xYy0yLjcgMC01LTIuMi01LTVWMTMuOWMuMS0yLjcgMi4zLTUgNS01eiIgY2xhc3M9InN0MiIvPjxwYXRoIGQ9Ik04OC43IDguOWgyNThjMi43IDAgNSAyLjIgNSA1djI4LjNjMCAyLjctMi4yIDUtNSA1aC0yNThjLTIuNyAwLTUtMi4yLTUtNVYxMy45YzAtMi43IDIuMi01IDUtNXoiIGNsYXNzPSJzdDMiLz48cGF0aCBkPSJNMzQ2LjcgODUuNWgtMjguM2MtMi43IDAtNSAyLjItNSA1djI4LjNjMCAyLjggMi4yIDUgNSA1aDI4LjNjMi43IDAgNS0yLjIgNS01VjkwLjVjMC0yLjgtMi4zLTUtNS01eiIgY2xhc3M9InN0MiIvPjxwYXRoIGQ9Ik0xOTMuNiA4NS41SDEyLjFjLTIuNyAwLTUgMi4zLTUgNXYyOC4zYzAgMi43IDIuMiA1IDUgNWgxODEuNWMyLjcgMCA1LTIuMiA1LTVWOTAuNWMwLTIuOC0yLjItNS01LTV6IiBjbGFzcz0ic3QzIi8+PHBhdGggZD0iTTI3MC4yIDg1LjVoLTI4LjNjLTIuNyAwLTUgMi4yLTUgNXYyOC4zYzAgMi44IDIuMiA1IDUgNWgyOC4zYzIuNyAwIDUtMi4yIDUtNVY5MC41Yy0uMS0yLjgtMi4zLTUtNS01eiIgY2xhc3M9InN0MiIvPjxwYXRoIGQ9Ik0yNzAuMiAxNjJIODguN2MtMi43IDAtNSAyLjItNSA1djI4LjNjMCAyLjcgMi4yIDUgNSA1aDE4MS41YzIuNyAwIDUtMi4yIDUtNVYxNjdjLS4xLTIuOC0yLjMtNS01LTV6IiBjbGFzcz0ic3QzIi8+PHBhdGggZD0iTTM0Ni43IDE2MmgtMjguM2MtMi43IDAtNSAyLjItNSA1djI4LjNjMCAyLjggMi4yIDUgNSA1aDI4LjNjMi43IDAgNS0yLjIgNS01VjE2N2MwLTIuOC0yLjMtNS01LTV6bS0zMDYuMyAwSDEyLjFjLTIuNyAwLTUgMi4yLTUgNXYyOC4zYzAgMi44IDIuMiA1IDUgNWgyOC4zYzIuNyAwIDUtMi4yIDUtNVYxNjdjMC0yLjgtMi4yLTUtNS01em0tMjguMyA3Ni41aDI4LjNjMi43IDAgNSAyLjIgNSA1djI4LjNjMCAyLjctMi4yIDUtNSA1SDEyLjFjLTIuNyAwLTUtMi4yLTUtNXYtMjguM2MuMS0yLjcgMi4zLTUgNS01eiIgY2xhc3M9InN0MiIvPjxwYXRoIGQ9Ik0xNjUuMiAyMzguNWgxODEuNWMyLjcgMCA1IDIuMiA1IDV2MjguM2MwIDIuNy0yLjIgNS01IDVIMTY1LjJjLTIuNyAwLTUtMi4yLTUtNXYtMjguM2MwLTIuNyAyLjItNSA1LTV6IiBjbGFzcz0ic3QzIi8+PHBhdGggZD0iTTg4LjcgMjM4LjVIMTE3YzIuNyAwIDUgMi4yIDUgNXYyOC4zYzAgMi43LTIuMiA1LTUgNUg4OC43Yy0yLjcgMC01LTIuMi01LTV2LTI4LjNjMC0yLjcgMi4yLTUgNS01em0yNTggNzYuN2gtMjguM2MtMi43IDAtNSAyLjItNSA1djI4LjNjMCAyLjggMi4yIDUgNSA1aDI4LjNjMi43IDAgNS0yLjIgNS01di0yOC4zYzAtMi44LTIuMy01LTUtNXoiIGNsYXNzPSJzdDIiLz48cGF0aCBkPSJNMjcwLjIgMzE1LjJoLTI1OGMtMi43IDAtNSAyLjItNSA1djI4LjNjMCAyLjcgMi4yIDUgNSA1aDI1OGMyLjcgMCA1LTIuMiA1LTV2LTI4LjNjLS4xLTIuOC0yLjMtNS01LTV6IiBjbGFzcz0ic3QzIi8+PC9nPjwvc3ZnPg==)](https://github.com/spiffe/spiffe/blob/main/MATURITY.md#development)

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
provides the Workload API implementation (e.g SPIRE agent) and registered
with the kubelet using the official CSI Node Driver Registrar image. The
SPIFFE CSI Driver and the Workload API implementation share the directory
hosting the Workload API Unix Domain Socket using an `emptyDir` mount.

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

### Failure to mount the socket directory

This problem can be diagnosed by dumping the SPIFFE CSI driver logs.

### Failure to Terminate Pods when Driver is Unhealthy Or Removed

If the SPIFFE CSI Driver is removed (or is otherwise unhealthy), any pods that
contain a volume mounted by the driver will fail to fully terminate until
driver health is restored. The describe command (i.e. kubectl describe) will
show the failure to unmount the volume. Kubernetes will continue to retry to
unmount the volume via the CSI driver. Once the driver has been restored, the
unmounting will eventually succeed and the pod will be fully terminated.

## Reporting a Vulnerability

Vulnerabilities can be reported by sending an email to security@spiffe.io. A
confirmation email will be sent to acknowledge the report within 72 hours. A
second acknowledgement will be sent within 7 days when the vulnerability has
been positively or negatively confirmed.
