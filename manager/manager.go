package manager

import (
	"context"
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ava-labs/vm-tester/constants"
	"github.com/ava-labs/vm-tester/utils"

	"github.com/ava-labs/avalanchego/api/info"
	"github.com/ava-labs/avalanchego/app/process"
	"github.com/ava-labs/avalanchego/node"
	"github.com/fatih/color"
	"golang.org/x/sync/errgroup"
)

const (
	bootstrapID = "NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg"
	bootstrapIP = "127.0.0.1:9651"
)

// Embed certs in binary and write to tmp file on startup (full binary)
var (
	//go:embed certs/keys1/staker.crt
	keys1StakerCrt []byte
	//go:embed certs/keys1/staker.key
	keys1StakerKey []byte

	//go:embed certs/keys2/staker.crt
	keys2StakerCrt []byte
	//go:embed certs/keys2/staker.key
	keys2StakerKey []byte

	//go:embed certs/keys3/staker.crt
	keys3StakerCrt []byte
	//go:embed certs/keys3/staker.key
	keys3StakerKey []byte

	//go:embed certs/keys4/staker.crt
	keys4StakerCrt []byte
	//go:embed certs/keys4/staker.key
	keys4StakerKey []byte

	//go:embed certs/keys5/staker.crt
	keys5StakerCrt []byte
	//go:embed certs/keys5/staker.key
	keys5StakerKey []byte

	nodeCerts = [][]byte{keys1StakerCrt, keys2StakerCrt, keys3StakerCrt, keys4StakerCrt, keys5StakerCrt}
	nodeKeys  = [][]byte{keys1StakerKey, keys2StakerKey, keys3StakerKey, keys4StakerKey, keys5StakerKey}
)

func NodeIDs() []string {
	nodeCerts := [][]byte{keys1StakerCrt, keys2StakerCrt, keys3StakerCrt, keys4StakerCrt, keys5StakerCrt}
	nodeIDs := make([]string, constants.NumNodes)
	for i, cert := range nodeCerts {
		id, err := utils.LoadNodeID(cert)
		if err != nil {
			panic(err)
		}
		nodeIDs[i] = id
	}
	return nodeIDs
}

func NodeURLs() []string {
	urls := make([]string, constants.NumNodes)
	for i := 0; i < constants.NumNodes; i++ {
		urls[i] = fmt.Sprintf("http://127.0.0.1:%d", constants.BaseHTTPPort+i*2)
	}
	return urls
}

func StartNetwork(ctx context.Context, vmPath string, bootstrapped chan struct{}) error {
	dir, err := ioutil.TempDir("", "vm-tester")
	if err != nil {
		panic(err)
	}
	color.Cyan("tmp dir located at: %s", dir)
	defer func() {
		color.Cyan("tmp dir located at: %s", dir)
	}()

	// Copy files into custom plugins
	pluginsDir := fmt.Sprintf("%s/plugins", dir)
	if err := os.MkdirAll(pluginsDir, os.FileMode(constants.FilePerms)); err != nil {
		panic(err)
	}
	if err := utils.CopyFile("system-plugins/evm", fmt.Sprintf("%s/evm", pluginsDir)); err != nil {
		panic(err)
	}
	if len(vmPath) > 0 {
		if err := utils.CopyFile(vmPath, fmt.Sprintf("%s/%s", pluginsDir, constants.VMID)); err != nil {
			panic(err)
		}
	}

	nodeConfigs := make([]node.Config, constants.NumNodes)
	for i := 0; i < constants.NumNodes; i++ {
		nodeDir := fmt.Sprintf("%s/node%d", dir, i+1)
		if err := os.MkdirAll(nodeDir, os.FileMode(constants.FilePerms)); err != nil {
			panic(err)
		}
		certFile := fmt.Sprintf("%s/staker.crt", nodeDir)
		if err := ioutil.WriteFile(certFile, nodeCerts[i], os.FileMode(constants.FilePerms)); err != nil {
			panic(err)
		}
		keyFile := fmt.Sprintf("%s/staker.key", nodeDir)
		if err := ioutil.WriteFile(keyFile, nodeKeys[i], os.FileMode(constants.FilePerms)); err != nil {
			panic(err)
		}

		df := defaultFlags()
		df.LogLevel = "info"
		df.LogDir = fmt.Sprintf("%s/logs", nodeDir)
		df.DBDir = fmt.Sprintf("%s/db", nodeDir)
		df.StakingEnabled = true
		df.HTTPPort = uint(constants.BaseHTTPPort + 2*i)
		df.StakingPort = uint(constants.BaseHTTPPort + 2*i + 1)
		if i != 0 {
			df.BootstrapIPs = bootstrapIP
			df.BootstrapIDs = bootstrapID
		} else {
			df.BootstrapIPs = ""
			df.BootstrapIDs = ""
		}
		if len(vmPath) > 0 {
			df.WhitelistedSubnets = constants.WhitelistedSubnets
		}
		df.StakingTLSCertFile = certFile
		df.StakingTLSKeyFile = keyFile
		nodeConfig, err := createNodeConfig(pluginsDir, flagsToArgs(df))
		if err != nil {
			panic(err)
		}
		nodeConfig.PluginDir = pluginsDir
		nodeConfigs[i] = nodeConfig
	}

	// Start all nodes and check if bootstrapped
	g, gctx := errgroup.WithContext(ctx)
	for _, config := range nodeConfigs {
		c := config
		g.Go(func() error {
			return runApp(g, gctx, c)
		})
	}
	g.Go(func() error {
		return checkBootstrapped(gctx, bootstrapped)
	})
	return g.Wait()
}

func checkBootstrapped(ctx context.Context, bootstrapped chan struct{}) error {
	if bootstrapped == nil {
		return nil
	}

	var (
		nodeURLs = NodeURLs()
		nodeIDs  = NodeIDs()
	)

	for i, url := range nodeURLs {
		client := info.NewClient(url, constants.HTTPTimeout)
		for {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			for _, chain := range constants.Chains {
				chainBootstrapped, _ := client.IsBootstrapped(chain)
				if !chainBootstrapped {
					time.Sleep(1 * time.Second)
					continue
				}
			}
			color.Cyan("%d is bootstrapped", nodeIDs[i])
			break
		}
	}

	color.Cyan("all nodes bootstrapped")
	close(bootstrapped)
	return nil
}

func runApp(g *errgroup.Group, ctx context.Context, config node.Config) error {
	app := process.NewApp(config)

	g.Go(func() error {
		<-ctx.Done()
		app.Stop()
		return nil
	})

	// Start running the AvalancheGo application
	exitCode := app.Start()
	return fmt.Errorf("unable to start: %d", exitCode)
}
