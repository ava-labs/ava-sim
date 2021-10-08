package manager

import (
	"context"
	_ "embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ava-labs/avalanchego/app/process"
	"github.com/ava-labs/avalanchego/node"
	"golang.org/x/sync/errgroup"
)

const (
	bootstrapID  = "NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg"
	bootstrapIP  = "127.0.0.1:9651"
	numNodes     = 5
	baseHTTPPort = 9650
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
)

func StartNetwork(ctx context.Context, pluginDir string, whitelistedSubnets string) error {
	dir, err := ioutil.TempDir("", "vm-tester")
	if err != nil {
		panic(err)
	}
	log.Println("created tmp dir", dir)
	// defer os.RemoveAll(dir)

	nodeConfigs := make([]node.Config, numNodes)
	nodeCerts := [][]byte{keys1StakerCrt, keys2StakerCrt, keys3StakerCrt, keys4StakerCrt, keys5StakerCrt}
	nodeKeys := [][]byte{keys1StakerKey, keys2StakerKey, keys3StakerKey, keys4StakerKey, keys5StakerKey}
	for i := 0; i < numNodes; i++ {
		nodeDir := fmt.Sprintf("%s/node%d", dir, i+1)
		if err := os.MkdirAll(nodeDir, os.FileMode(0777)); err != nil {
			panic(err)
		}
		certFile := fmt.Sprintf("%s/staker.crt", nodeDir)
		if err := ioutil.WriteFile(certFile, nodeCerts[i], os.FileMode(0777)); err != nil {
			panic(err)
		}
		keyFile := fmt.Sprintf("%s/staker.key", nodeDir)
		if err := ioutil.WriteFile(keyFile, nodeKeys[i], os.FileMode(0777)); err != nil {
			panic(err)
		}

		df := defaultFlags()
		df.LogLevel = "info"
		df.LogDir = fmt.Sprintf("%s/logs", nodeDir)
		df.DBDir = fmt.Sprintf("%s/db", nodeDir)
		df.StakingEnabled = true
		df.HTTPPort = uint(baseHTTPPort + 2*i)
		df.StakingPort = uint(baseHTTPPort + 2*i + 1)
		if i != 0 {
			df.BootstrapIPs = bootstrapIP
			df.BootstrapIDs = bootstrapID
		} else {
			df.BootstrapIPs = ""
			df.BootstrapIDs = ""
		}
		df.WhitelistedSubnets = whitelistedSubnets
		df.StakingTLSCertFile = certFile
		df.StakingTLSKeyFile = keyFile
		df.WhitelistedSubnets = whitelistedSubnets
		fmt.Println(df)
		nodeConfig, err := createNodeConfig(pluginDir, flagsToArgs(df))
		if err != nil {
			panic(err)
		}
		nodeConfig.PluginDir = pluginDir
		nodeConfigs[i] = nodeConfig
	}

	g, gctx := errgroup.WithContext(ctx)
	for _, config := range nodeConfigs {
		c := config
		g.Go(func() error {
			return runApp(g, gctx, c)
		})
	}

	return g.Wait()
}

func runApp(g *errgroup.Group, ctx context.Context, config node.Config) error {
	app := process.NewApp(config)

	g.Go(func() error {
		<-ctx.Done()
		app.Stop()
		return nil
	})

	// start running the application
	exitCode := app.Start()
	return fmt.Errorf("unable to start: %d", exitCode)
}
