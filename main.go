package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/ava-labs/ava-sim/manager"
	"github.com/ava-labs/ava-sim/runner"
	"github.com/fatih/color"
	"golang.org/x/sync/errgroup"
)

func main() {
	var vm, vmGenesis string
	switch len(os.Args) {
	case 1: // normal network
	case 3:
		vm = path.Clean(os.Args[1])
		if _, err := os.Stat(vm); os.IsNotExist(err) {
			panic(fmt.Sprintf("%s does not exist", vm))
		}
		color.Yellow("vm set to: %s", vm)

		vmGenesis = path.Clean(os.Args[2])
		if _, err := os.Stat(vmGenesis); os.IsNotExist(err) {
			panic(fmt.Sprintf("%s does not exist", vmGenesis))
		}
		color.Yellow("vm-genesis set to: %s", vmGenesis)
	default:
		panic("invalid arguments (expecting no arguments or [vm] [vm-genesis])")
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
		case sig := <-signals:
			color.Red("signal received: %v", sig)
			cancel()
		case <-gctx.Done():
		}
		return nil
	})

	g.Go(func() error {
		return manager.StartNetwork(gctx, vm, bootstrapped)
	})

	// Only setup network if a custom VM is provided and the network has finished
	// bootstrapping
	select {
	case <-bootstrapped:
		if len(vm) > 0 && gctx.Err() == nil {
			g.Go(func() error {
				return runner.SetupSubnet(gctx, vmGenesis)
			})
		}
	case <-gctx.Done():
	}

	color.Red("ava-sim exited with error: %s", g.Wait())
	os.Exit(1)
}
