package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func main() {
	updateStatus("starting")
	workloadAPIAddr := flag.String("workload-api-addr", "", "Workload API Address")
	flag.Parse()

	var opts []workloadapi.ClientOption
	if *workloadAPIAddr != "" {
		opts = append(opts, workloadapi.WithAddr(*workloadAPIAddr))
	}

	log.Println("Watching...")
	err := workloadapi.WatchX509Context(context.Background(), watcher{}, opts...)
	log.Fatal("Error: ", err)
}

type watcher struct{}

func (watcher) OnX509ContextUpdate(x509Context *workloadapi.X509Context) {
	log.Println("X509 context updated.")
	updateStatus("updated")
}

func (watcher) OnX509ContextWatchError(err error) {
	log.Println("Error:", err)
	updateStatus("error")
}

func updateStatus(status string) {
	os.WriteFile("status", []byte(status+"\n"), 0644)
}
