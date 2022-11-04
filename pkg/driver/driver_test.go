package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/go-logr/logr"
	"github.com/spiffe/spiffe-csi/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	testNodeID = "nodeID"
)

func init() {
	bindMountRW = func(src, dst string) error {
		return writeMeta(dst, src)
	}
	unmount = func(dst string) error {
		return os.Remove(metaPath(dst))
	}
}

func TestNew(t *testing.T) {
	workloadAPISocketDir := t.TempDir()

	t.Run("node ID is required", func(t *testing.T) {
		_, err := New(Config{
			WorkloadAPISocketDir: workloadAPISocketDir,
		})
		require.EqualError(t, err, "node ID is required")
	})

	t.Run("workload API socket directory is required", func(t *testing.T) {
		_, err := New(Config{
			NodeID: testNodeID,
		})
		require.EqualError(t, err, "workload API socket directory is required")
	})

	t.Run("success", func(t *testing.T) {
		_, err := New(Config{
			NodeID:               testNodeID,
			WorkloadAPISocketDir: workloadAPISocketDir,
		})
		require.NoError(t, err)
	})
}

func TestBoilerplateRPCs(t *testing.T) {
	client, _ := startDriver(t)

	t.Run("GetPluginInfo", func(t *testing.T) {
		resp, err := client.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})
		require.NoError(t, err)
		requireProtoEqual(t, &csi.GetPluginInfoResponse{
			Name:          "csi.spiffe.io",
			VendorVersion: version.Version(),
		}, resp, "unexpected response")
	})

	t.Run("GetPluginCapabilities", func(t *testing.T) {
		resp, err := client.GetPluginCapabilities(context.Background(), &csi.GetPluginCapabilitiesRequest{})
		require.NoError(t, err)
		requireProtoEqual(t, &csi.GetPluginCapabilitiesResponse{}, resp, "unexpected response")
	})

	t.Run("Probe", func(t *testing.T) {
		resp, err := client.Probe(context.Background(), &csi.ProbeRequest{})
		require.NoError(t, err)
		requireProtoEqual(t, &csi.ProbeResponse{}, resp, "unexpected response")
	})

	t.Run("NodeGetCapabilities", func(t *testing.T) {
		resp, err := client.NodeGetCapabilities(context.Background(), &csi.NodeGetCapabilitiesRequest{})
		require.NoError(t, err)
		requireProtoEqual(t, &csi.NodeGetCapabilitiesResponse{
			Capabilities: []*csi.NodeServiceCapability{
				{
					Type: &csi.NodeServiceCapability_Rpc{
						Rpc: &csi.NodeServiceCapability_RPC{
							Type: csi.NodeServiceCapability_RPC_VOLUME_CONDITION,
						},
					},
				},
				{
					Type: &csi.NodeServiceCapability_Rpc{
						Rpc: &csi.NodeServiceCapability_RPC{
							Type: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
						},
					},
				},
			},
		}, resp, "unexpected response")
	})

	t.Run("NodeGetInfo", func(t *testing.T) {
		resp, err := client.NodeGetInfo(context.Background(), &csi.NodeGetInfoRequest{})
		require.NoError(t, err)
		requireProtoEqual(t, &csi.NodeGetInfoResponse{
			NodeId:            testNodeID,
			MaxVolumesPerNode: 0,
		}, resp, "unexpected response")
	})
}

func TestNodePublishVolume(t *testing.T) {
	for _, tt := range []struct {
		desc            string
		mutateReq       func(req *csi.NodePublishVolumeRequest)
		mungeTargetPath func(t *testing.T, targetPath string)
		expectCode      codes.Code
		expectMsgPrefix string
	}{
		{
			desc: "missing volume id",
			mutateReq: func(req *csi.NodePublishVolumeRequest) {
				req.VolumeId = ""
			},
			expectCode:      codes.InvalidArgument,
			expectMsgPrefix: "request missing required volume id",
		},
		{
			desc: "missing target path",
			mutateReq: func(req *csi.NodePublishVolumeRequest) {
				req.TargetPath = ""
			},
			expectCode:      codes.InvalidArgument,
			expectMsgPrefix: "request missing required target path",
		},
		{
			desc: "missing volume capability",
			mutateReq: func(req *csi.NodePublishVolumeRequest) {
				req.VolumeCapability = nil
			},
			expectCode:      codes.InvalidArgument,
			expectMsgPrefix: "request missing required volume capability",
		},
		{
			desc: "missing volume capability access type",
			mutateReq: func(req *csi.NodePublishVolumeRequest) {
				req.VolumeCapability.AccessType = nil
			},
			expectCode:      codes.InvalidArgument,
			expectMsgPrefix: "request missing required volume capability access type",
		},
		{
			desc: "invalid volume capability access type mount fs type",
			mutateReq: func(req *csi.NodePublishVolumeRequest) {
				req.VolumeCapability.AccessType = &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{
						FsType: "ANTHING HERE IS BAD",
					},
				}
			},
			expectCode:      codes.InvalidArgument,
			expectMsgPrefix: "request volume capability access type must be a simple mount",
		},
		{
			desc: "invalid volume capability access type mount fs type",
			mutateReq: func(req *csi.NodePublishVolumeRequest) {
				req.VolumeCapability.AccessType = &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{
						MountFlags: []string{"ANYTHING HERE IS BAD"},
					},
				}
			},
			expectCode:      codes.InvalidArgument,
			expectMsgPrefix: "request volume capability access type must be a simple mount",
		},
		{
			desc: "invalid volume capability access type",
			mutateReq: func(req *csi.NodePublishVolumeRequest) {
				req.VolumeCapability.AccessType = &csi.VolumeCapability_Block{}
			},
			expectCode:      codes.InvalidArgument,
			expectMsgPrefix: "request volume capability access type must be a simple mount",
		},
		{
			desc: "missing volume capability access mode",
			mutateReq: func(req *csi.NodePublishVolumeRequest) {
				req.VolumeCapability.AccessMode = nil
			},
			expectCode:      codes.InvalidArgument,
			expectMsgPrefix: "request missing required volume capability access mode",
		},
		{
			desc: "invalid volume capability access mode",
			mutateReq: func(req *csi.NodePublishVolumeRequest) {
				req.VolumeCapability.AccessMode.Mode = csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY
			},
			expectCode:      codes.InvalidArgument,
			expectMsgPrefix: "request volume capability access mode is not valid",
		},
		{
			desc: "not an ephemeral volume",
			mutateReq: func(req *csi.NodePublishVolumeRequest) {
				req.VolumeContext = nil
			},
			expectCode:      codes.InvalidArgument,
			expectMsgPrefix: "only ephemeral volumes are supported",
		},
		{
			desc: "target path already exists",
			mungeTargetPath: func(t *testing.T, targetPath string) {
				require.NoError(t, os.Mkdir(targetPath, 0755))
			},
			expectCode: codes.OK,
		},
		{
			desc: "unable to create target path when missing",
			mungeTargetPath: func(t *testing.T, targetPath string) {
				// By removing the parent directory, the code that attempts to
				// create the target path directory will fail (it assumes it
				// will exist).
				require.NoError(t, os.Remove(filepath.Dir(targetPath)))
			},
			expectCode:      codes.Internal,
			expectMsgPrefix: "unable to create target path",
		},
		{
			desc: "mount failure",
			mungeTargetPath: func(t *testing.T, targetPath string) {
				// write out a file to the target path... this will prevent our
				// fake mount implementation from being able to write the
				// metadata file, thus simulating a mount failure.
				require.NoError(t, os.WriteFile(targetPath, nil, 0600))
			},
			expectCode:      codes.Internal,
			expectMsgPrefix: "unable to mount",
		},
		{
			desc: "enforcing read-only volumes",
			mutateReq: func(req *csi.NodePublishVolumeRequest) {
				req.Readonly = false
			},
			expectCode:      codes.InvalidArgument,
			expectMsgPrefix: "pod.spec.volumes[].csi.readOnly must be set to 'true'",
		},
		{
			desc:       "success",
			expectCode: codes.OK,
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			targetPathBase := t.TempDir()
			targetPath := filepath.Join(targetPathBase, "target-path")

			if tt.mungeTargetPath != nil {
				tt.mungeTargetPath(t, targetPath)
			}

			req := &csi.NodePublishVolumeRequest{
				VolumeId:   "volumeID",
				TargetPath: targetPath,
				Readonly:   true,
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{},
					AccessMode: &csi.VolumeCapability_AccessMode{},
				},
				VolumeContext: map[string]string{
					"csi.storage.k8s.io/ephemeral": "true",
				},
			}
			if tt.mutateReq != nil {
				tt.mutateReq(req)
			}

			client, workloadAPISocketDir := startDriver(t)

			resp, err := client.NodePublishVolume(context.Background(), req)
			requireGRPCStatusPrefix(t, err, tt.expectCode, tt.expectMsgPrefix)
			if err == nil {
				assert.Equal(t, &csi.NodePublishVolumeResponse{}, resp)
				assertMounted(t, targetPath, workloadAPISocketDir)
			} else {
				assert.Nil(t, resp)
				assertNotMounted(t, targetPath)
			}
		})
	}
}

func TestNodeUnpublishVolume(t *testing.T) {
	client, workloadAPISocketDir := startDriver(t)

	for _, tt := range []struct {
		desc            string
		mutateReq       func(req *csi.NodeUnpublishVolumeRequest)
		mungeTargetPath func(t *testing.T, targetPath string)
		expectCode      codes.Code
		expectMsgPrefix string
	}{
		{
			desc: "missing volume id",
			mutateReq: func(req *csi.NodeUnpublishVolumeRequest) {
				req.VolumeId = ""
			},
			expectCode:      codes.InvalidArgument,
			expectMsgPrefix: "request missing required volume id",
		},
		{
			desc: "missing target path",
			mutateReq: func(req *csi.NodeUnpublishVolumeRequest) {
				req.TargetPath = ""
			},
			expectCode:      codes.InvalidArgument,
			expectMsgPrefix: "request missing required target path",
		},
		{
			desc: "unmount failure",
			mungeTargetPath: func(t *testing.T, targetPath string) {
				// Removing the meta file to simulate that it wasn't mounted
				require.NoError(t, os.Remove(metaPath(targetPath)))
			},
			expectCode:      codes.Internal,
			expectMsgPrefix: "unable to unmount",
		},
		{
			desc: "unable to remove target path after unmounting",
			mungeTargetPath: func(t *testing.T, targetPath string) {
				// Prevent the directory from being removed by writing
				// a file into it.
				require.NoError(t, os.WriteFile(filepath.Join(targetPath, "prevent-directory-removal"), nil, 0600))
			},
			expectCode:      codes.Internal,
			expectMsgPrefix: "unable to remove target path",
		},
		{
			desc:       "success",
			expectCode: codes.OK,
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			targetPathBase := t.TempDir()
			targetPath := filepath.Join(targetPathBase, "target-path")

			// Write out the meta file to simulate a successful mount
			require.NoError(t, os.Mkdir(targetPath, 0755))
			require.NoError(t, writeMeta(targetPath, workloadAPISocketDir))

			if tt.mungeTargetPath != nil {
				tt.mungeTargetPath(t, targetPath)
			}

			req := &csi.NodeUnpublishVolumeRequest{
				VolumeId:   "volumeID",
				TargetPath: targetPath,
			}
			if tt.mutateReq != nil {
				tt.mutateReq(req)
			}
			dumpIt(t, "BEFORE", targetPathBase)
			resp, err := client.NodeUnpublishVolume(context.Background(), req)
			dumpIt(t, "AFTER", targetPathBase)
			requireGRPCStatusPrefix(t, err, tt.expectCode, tt.expectMsgPrefix)
			if err == nil {
				assertNotMounted(t, targetPath)
				assert.Equal(t, &csi.NodeUnpublishVolumeResponse{}, resp)
			} else {
				assert.Nil(t, resp)
			}
		})
	}
}

func requireGRPCStatusPrefix(tb testing.TB, err error, code codes.Code, msgPrefix string, msgAndArgs ...interface{}) {
	st := status.Convert(err)
	if code != st.Code() || !strings.HasPrefix(st.Message(), msgPrefix) {
		require.Fail(tb, fmt.Sprintf("Status code=%q msg=%q does not match code=%q with msg prefix %q", st.Code(), st.Message(), code, msgPrefix), msgAndArgs...)
	}
}

func requireProtoEqual(tb testing.TB, expected, actual interface{}, msgAndArgs ...interface{}) {
	// The CSI spec codegen uses the old proto package, which doesn't have
	// good comparison support, so just render the structs as json and
	// compare.
	expectedJSON, err := json.Marshal(expected)
	require.NoError(tb, err, "expected cannot be marshaled to JSON")
	actualJSON, err := json.Marshal(actual)
	require.NoError(tb, err, "actual cannot be marshaled to JSON")
	require.JSONEq(tb, string(expectedJSON), string(actualJSON), msgAndArgs...)
}

type client struct {
	csi.IdentityClient
	csi.NodeClient
}

func startDriver(t *testing.T) (client, string) {
	workloadAPISocketDir := t.TempDir()

	d, err := New(Config{
		Log:                  logr.Discard(),
		NodeID:               testNodeID,
		WorkloadAPISocketDir: workloadAPISocketDir,
	})
	require.NoError(t, err)

	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	t.Cleanup(func() { l.Close() })
	s := grpc.NewServer()
	t.Cleanup(s.Stop)

	csi.RegisterIdentityServer(s, d)
	csi.RegisterNodeServer(s, d)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	connCh := make(chan *grpc.ClientConn, 1)
	errCh := make(chan error, 2)

	go func() {
		errCh <- s.Serve(l) // failures to serve will
	}()
	go func() {
		conn, err := grpc.DialContext(ctx, l.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.FailOnNonTempDialError(true),
			grpc.WithReturnConnectionError())
		if err != nil {
			errCh <- err
		} else {
			connCh <- conn
		}
	}()

	var conn *grpc.ClientConn
	select {
	case conn = <-connCh:
		t.Cleanup(func() { conn.Close() })
	case err := <-errCh:
		require.NoError(t, err)
	}

	return client{
		IdentityClient: csi.NewIdentityClient(conn),
		NodeClient:     csi.NewNodeClient(conn),
	}, workloadAPISocketDir
}

func assertMounted(t *testing.T, targetPath, src string) {
	meta, err := readMeta(targetPath)
	if assert.NoError(t, err) {
		assert.Equal(t, src, meta)
	}
}

func assertNotMounted(t *testing.T, targetPath string) {
	_, err := readMeta(targetPath)
	assert.Error(t, err, "should not be mounted")
}

func readMeta(targetPath string) (string, error) {
	data, err := os.ReadFile(metaPath(targetPath))
	return string(data), err
}

func writeMeta(targetPath string, meta string) error {
	return os.WriteFile(metaPath(targetPath), []byte(meta), 0600)
}

func metaPath(targetPath string) string {
	return filepath.Join(targetPath, "meta")
}

func dumpIt(t *testing.T, when, dir string) {
	t.Logf(">>>>>>>>>> DUMPING %s %s", when, dir)
	assert.NoError(t, filepath.Walk(dir, filepath.WalkFunc(
		func(path string, info fs.FileInfo, err error) error {
			t.Logf("%s: %o", path, info.Mode())
			return nil
		})))
	t.Logf("<<<<<<<<<< DUMPED %s %s", when, dir)
}
