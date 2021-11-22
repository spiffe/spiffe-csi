package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/spiffe/spiffe-csi/internal/version"
	"go.uber.org/zap"
)

var (
	nodeIDFlag           = flag.String("node-id", "", "Kubernetes Node ID. If unset, the node ID is obtained from the environment (i.e., -node-id-env)")
	nodeIDEnvFlag        = flag.String("node-id-env", "MY_NODE_NAME", "Envvar from which to obtain the node ID. Overriden by -node-id.")
	csiSocketPathFlag    = flag.String("csi-socket-path", "/spiffe-csi/csi.sock", "Path to the CSI socket")
	workloadAPISocketDir = flag.String("workload-api-socket-dir", "", "Path to the Workload API socket directory")
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
		fmt.Fprintf(os.Stderr, "Unable to set up logger: %w", err)
		os.Exit(1)
	}
	log = zapr.NewLogger(zapLog)

	config := Config{
		Log:                  log,
		NodeID:               getNodeIDFromFlags(),
		WorkloadAPISocketDir: *workloadAPISocketDir,
		CSISocketPath:        *csiSocketPathFlag,
	}

	log.Info("Starting.",
		"version", version.Version(),
		"node-id", config.NodeID,
		"workload-api-socket-dir", config.WorkloadAPISocketDir,
		"csi-socket-path", config.CSISocketPath,
	)
	if err := Run(config); err != nil {
		log.Error(err, "Failed to run")
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
