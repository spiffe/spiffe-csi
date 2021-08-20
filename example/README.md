# SPIFFE CSI Driver Example

TODO: fill this in with more detail

1. build CSI driver image via project Makefile (then tag as test)
1. build example workload image via workload/Makefile
1. start kind cluster
1. load images into kind cluster (the test tags for both the CSI driver and workload)
1. create spire namespace
1. apply the following yaml
   * spiffe-csi-driver.yaml
   * spire-server.yaml
   * spire-agent.yaml
   * workload.yaml
1. create a spire entry for the workload
1. check workload logs
