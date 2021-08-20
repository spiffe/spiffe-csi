package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/spiffe/spiffe-csi/internal/version"
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

	log.Println("Starting...")

	config := Config{
		NodeID:               getNodeIDFromFlags(),
		WorkloadAPISocketDir: *workloadAPISocketDir,
		CSISocketPath:        *csiSocketPathFlag,
	}

	log.Println("Version              :", version.Version())
	log.Println("Node ID              :", config.NodeID)
	log.Println("WorkloadAPISocketDir :", config.WorkloadAPISocketDir)
	log.Println("CSISocketPath        :", config.CSISocketPath)

	if err := Run(config); err != nil {
		log.Fatal("Error:", err)
	}
	log.Println("Done")
}

func getNodeIDFromFlags() string {
	nodeID := os.Getenv(*nodeIDEnvFlag)
	if *nodeIDFlag != "" {
		nodeID = *nodeIDFlag
	}
	return nodeID
}
