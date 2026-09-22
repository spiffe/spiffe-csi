// Package main is the entry point for the SPIFFE CSI driver.
package main

import (
	"flag"
	"fmt"
	"os"

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
	pluginNameFlag           = flag.String("plugin-name", "csi.spiffe.io", "Plugin name to register")
	workloadAPISocketDirFlag = flag.String("workload-api-socket-dir", "", "Path to the Workload API socket directory")
	logFormatFlag            = flag.String("log-format", "text", "Log format: text or json")
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

	zapLog, err := newZapLogger(*logFormatFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to set up logger: %v\n", err)
		os.Exit(1)
	}
	log := zapr.NewLogger(zapLog)

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
		PluginName:           *pluginNameFlag,
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

func newZapLogger(format string) (*zap.Logger, error) {
	switch format {
	case "text":
		return zap.NewDevelopment()
	case "json":
		return zap.NewProduction()
	default:
		return nil, fmt.Errorf("invalid log-format %q: must be text or json", format)
	}
}

func getNodeIDFromFlags() string {
	nodeID := os.Getenv(*nodeIDEnvFlag)
	if *nodeIDFlag != "" {
		nodeID = *nodeIDFlag
	}
	return nodeID
}
