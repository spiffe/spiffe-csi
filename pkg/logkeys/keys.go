// Package logkeys defines structured log field keys used across the CSI driver.
package logkeys

// Log field keys for structured logging.
const (
	CSISocketPath        = "csiSocketPath"
	FullMethod           = "fullMethod"
	NodeID               = "nodeID"
	TargetPath           = "targetPath"
	Version              = "version"
	VolumeID             = "volumeID"
	VolumePath           = "volumePath"
	WorkloadAPISocketDir = "workloadAPISocketDir"
)
