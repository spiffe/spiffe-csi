package driver

import (
	"context"
	"errors"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/go-logr/logr"
	"github.com/spiffe/spiffe-csi/internal/version"
	"github.com/spiffe/spiffe-csi/pkg/logkeys"
	"github.com/spiffe/spiffe-csi/pkg/mount"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	pluginName = "csi.spiffe.io"
)

var (
	// We replace these in tests since bind mounting generally requires root.
	bindMountRO = mount.BindMountRO
	unmount     = mount.Unmount
)

// Config is the configuration for the driver
type Config struct {
	Log                  logr.Logger
	NodeID               string
	WorkloadAPISocketDir string
}

// Driver is the ephemeral-inline CSI driver implementation
type Driver struct {
	csi.UnimplementedIdentityServer
	csi.UnimplementedNodeServer

	log                  logr.Logger
	nodeID               string
	workloadAPISocketDir string
}

// New creates a new driver with the given config
func New(config Config) (*Driver, error) {
	switch {
	case config.NodeID == "":
		return nil, errors.New("node ID is required")
	case config.WorkloadAPISocketDir == "":
		return nil, errors.New("workload API socket directory is required")
	}
	return &Driver{
		log:                  config.Log,
		nodeID:               config.NodeID,
		workloadAPISocketDir: config.WorkloadAPISocketDir,
	}, nil
}

/////////////////////////////////////////////////////////////////////////////
// Identity Server
/////////////////////////////////////////////////////////////////////////////

func (d *Driver) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return &csi.GetPluginInfoResponse{
		Name:          pluginName,
		VendorVersion: version.Version(),
	}, nil
}

func (d *Driver) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	// Only the Node server is implemented. No other capabilities are available.
	return &csi.GetPluginCapabilitiesResponse{}, nil
}

func (d *Driver) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}

/////////////////////////////////////////////////////////////////////////////
// Node Server implementation
/////////////////////////////////////////////////////////////////////////////

func (d *Driver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (_ *csi.NodePublishVolumeResponse, err error) {
	ephemeralMode := req.GetVolumeContext()["csi.storage.k8s.io/ephemeral"]

	log := d.log.WithValues(
		logkeys.VolumeID, req.VolumeId,
		logkeys.TargetPath, req.TargetPath,
	)
	if req.VolumeCapability != nil && req.VolumeCapability.AccessMode != nil {
		log = log.WithValues("access_mode", req.VolumeCapability.AccessMode.Mode)
	}

	defer func() {
		if err != nil {
			log.Error(err, "Failed to publish volume")
		}
	}()

	// Validate request
	switch {
	case req.VolumeId == "":
		return nil, status.Error(codes.InvalidArgument, "request missing required volume id")
	case req.TargetPath == "":
		return nil, status.Error(codes.InvalidArgument, "request missing required target path")
	case req.VolumeCapability == nil:
		return nil, status.Error(codes.InvalidArgument, "request missing required volume capability")
	case req.VolumeCapability.AccessType == nil:
		return nil, status.Error(codes.InvalidArgument, "request missing required volume capability access type")
	case !isVolumeCapabilityPlainMount(req.VolumeCapability):
		return nil, status.Error(codes.InvalidArgument, "request volume capability access type must be a simple mount")
	case req.VolumeCapability.AccessMode == nil:
		return nil, status.Error(codes.InvalidArgument, "request missing required volume capability access mode")
	case isVolumeCapabilityAccessModeReadOnly(req.VolumeCapability.AccessMode):
		return nil, status.Error(codes.InvalidArgument, "request volume capability access mode is not valid")
	case ephemeralMode != "true":
		return nil, status.Error(codes.InvalidArgument, "only ephemeral volumes are supported")
	}

	// Create the target path (required by CSI interface)
	if err := os.Mkdir(req.TargetPath, 0777); err != nil && !os.IsExist(err) {
		return nil, status.Errorf(codes.Internal, "unable to create target path %q: %v", req.TargetPath, err)
	}
	// Bind mount the agent socket directory (read only) to the target path
	if err := bindMountRO(d.workloadAPISocketDir, req.TargetPath); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to mount %q: %v", req.TargetPath, err)
	}

	log.Info("Volume published")

	return &csi.NodePublishVolumeResponse{}, nil
}

func (d *Driver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (_ *csi.NodeUnpublishVolumeResponse, err error) {
	log := d.log.WithValues(
		logkeys.VolumeID, req.VolumeId,
		logkeys.TargetPath, req.TargetPath,
	)

	defer func() {
		if err != nil {
			log.Error(err, "Failed to unpublish volume")
		}
	}()

	// Validate request
	switch {
	case req.VolumeId == "":
		return nil, status.Error(codes.InvalidArgument, "request missing required volume id")
	case req.TargetPath == "":
		return nil, status.Error(codes.InvalidArgument, "request missing required target path")
	}

	if err := unmount(req.TargetPath); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to unmount %q: %v", req.TargetPath, err)
	}
	if err := os.Remove(req.TargetPath); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to remove target path %q: %v", req.TargetPath, err)
	}

	log.Info("Volume unpublished")

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (d *Driver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{}, nil
}

func (d *Driver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId:            d.nodeID,
		MaxVolumesPerNode: 0,
	}, nil
}

func isVolumeCapabilityPlainMount(volumeCapability *csi.VolumeCapability) bool {
	mount := volumeCapability.GetMount()
	switch {
	case mount == nil:
		return false
	case mount.FsType != "":
		return false
	case len(mount.MountFlags) != 0:
		return false
	}
	return true
}

func isVolumeCapabilityAccessModeReadOnly(accessMode *csi.VolumeCapability_AccessMode) bool {
	return accessMode.Mode == csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY
}
