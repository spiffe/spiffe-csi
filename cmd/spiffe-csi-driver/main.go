package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/spiffe/spiffe-csi/internal/version"
	"github.com/spiffe/spiffe-csi/pkg/driver"
	"github.com/spiffe/spiffe-csi/pkg/logkeys"
	"github.com/spiffe/spiffe-csi/pkg/server"
	"go.uber.org/zap"
)

var (
	nodeIDFlag               = flag.String("node-id", "", "Kubernetes Node ID. If unset, the node ID is obtained from the environment (i.e., -node-id-env)")
	nodeIDEnvFlag            = flag.String("node-id-env", "MY_NODE_NAME", "Envvar from which to obtain the node ID. Overridden by -node-id.")
	csiSocketPathFlag        = flag.String("csi-socket-path", "/spiffe-csi/csi.sock", "Path to the CSI socket")
	workloadAPISocketDirFlag = flag.String("workload-api-socket-dir", "", "Path to the Workload API socket directory")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s (version %s)\n", "spiffe-csi-driver", version.Version())
		fmt.Fprintln(os.Stderr, "Provides the Workload API socket directory via ephemeral inline CSI volumes")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	var log logr.Logger
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to set up logger: %v", err)
		os.Exit(1)
	}
	log = zapr.NewLogger(zapLog)

	nodeID := getNodeIDFromFlags()

	log.Info("Starting.",
		logkeys.Version, version.Version(),
		logkeys.NodeID, nodeID,
		logkeys.WorkloadAPISocketDir, *workloadAPISocketDirFlag,
		logkeys.CSISocketPath, *csiSocketPathFlag,
	)

	driver, err := driver.New(driver.Config{
		Log:                  log,
		NodeID:               nodeID,
		WorkloadAPISocketDir: *workloadAPISocketDirFlag,
	})
	if err != nil {
		log.Error(err, "Failed to create driver")
		os.Exit(1)
	}

	serverConfig := server.Config{
		Log:           log,
		CSISocketPath: *csiSocketPathFlag,
		Driver:        driver,
	}

	if err := server.Run(serverConfig); err != nil {
		log.Error(err, "Failed to serve")
		os.Exit(1)
	}
	log.Info("Done")
}

func getNodeIDFromFlags() string {
	nodeID := os.Getenv(*nodeIDEnvFlag)
	if *nodeIDFlag != "" {
		nodeID = *nodeIDFlag
	}
	return nodeID
}
