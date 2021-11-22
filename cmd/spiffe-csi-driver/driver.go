package main

import (
	"context"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/go-logr/logr"
	"github.com/spiffe/spiffe-csi/internal/version"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	pluginName = "csi.spiffe.io"
)

type Driver struct {
	csi.UnimplementedIdentityServer
	csi.UnimplementedNodeServer

	Log                  logr.Logger
	NodeID               string
	WorkloadAPISocketDir string
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

	log := d.Log.WithValues(
		"request-volume-id", req.VolumeId,
		"request-target-path", req.TargetPath,
	)

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
		return nil, status.Error(codes.InvalidArgument, "request volume capability access type must be mount")
	case req.VolumeCapability.AccessMode == nil:
		return nil, status.Error(codes.InvalidArgument, "request missing required volume capability access mode")
	case isVolumeCapabilityAccessModeReadOnly(req.VolumeCapability.AccessMode):
		return nil, status.Error(codes.InvalidArgument, "request missing required volume capability access mode")
	case ephemeralMode != "true":
		return nil, status.Error(codes.InvalidArgument, "only ephemeral volumes are supported")
	}

	// Create the target path (required by CSI interface)
	if err := os.Mkdir(req.TargetPath, 0777); err != nil && !os.IsExist(err) {
		return nil, status.Errorf(codes.Internal, "unable to create target path %q: %v", req.TargetPath, err)
	}
	// Bind mount the agent socket directory (read only) to the target path
	if err := bindMountRO(d.WorkloadAPISocketDir, req.TargetPath); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to mount %q: %v", req.TargetPath, err)
	}

	log.Info("Volume published")

	return &csi.NodePublishVolumeResponse{}, nil
}

func (d *Driver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (_ *csi.NodeUnpublishVolumeResponse, err error) {
	log := d.Log.WithValues(
		"request-volume-id", req.VolumeId,
		"request-target-path", req.TargetPath,
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
		NodeId:            d.NodeID,
		MaxVolumesPerNode: 1,
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
