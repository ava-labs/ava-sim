package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/ava-labs/vm-tester/manager"
	"github.com/ava-labs/vm-tester/runner"
	"golang.org/x/sync/errgroup"
)

func main() {
	// Parse Args
	rawConfigDir := flag.String("config-dir", "", "directory for all VM configs")
	rawVMPath := flag.String("vm-path", "", "location of custom VM binary")
	rawVMGenesis := flag.String("vm-genesis", "", "location of custom VM genesis")
	flag.Parse()
	var configDir, vmPath, vmGenesis string
	// TODO: use config path for custom VM only
	if len(*rawConfigDir) > 1 {
		configDir = path.Clean(*rawConfigDir)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			panic(fmt.Sprintf("%s does not exist", configDir))
		}
	}
	if len(*rawVMPath) > 1 {
		vmPath = path.Clean(*rawVMPath)
		if _, err := os.Stat(vmPath); os.IsNotExist(err) {
			panic(fmt.Sprintf("%s does not exist", vmPath))
		}
	}
	if len(*rawVMGenesis) > 1 {
		vmGenesis = path.Clean(*rawVMGenesis)
		if _, err := os.Stat(vmGenesis); os.IsNotExist(err) {
			panic(fmt.Sprintf("%s does not exist", vmGenesis))
		}
	}

	// Start local network
	bootstrapped := make(chan struct{})
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		// register signals to kill the application
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT)
		signal.Notify(signals, syscall.SIGTERM)
		defer func() {
			// shut down the signal go routine
			signal.Stop(signals)
			close(signals)
		}()

		select {
		case <-signals:
			cancel()
		case <-gctx.Done():
		}
		return nil
	})

	g.Go(func() error {
		return manager.StartNetwork(gctx, configDir, vmPath, bootstrapped)
	})
	<-bootstrapped

	// only setup network if customVM exists
	if len(vmPath) > 0 {
		g.Go(func() error {
			return runner.SetupSubnet(gctx, vmGenesis)
		})
	}
	log.Fatal(g.Wait())
}
