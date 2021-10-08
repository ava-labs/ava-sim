package manager

import (
	"context"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/ava-labs/avalanchego/ids"
	aConstants "github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/hashing"

	"github.com/ava-labs/vm-tester/constants"

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

	nodeCerts = [][]byte{keys1StakerCrt, keys2StakerCrt, keys3StakerCrt, keys4StakerCrt, keys5StakerCrt}
	nodeKeys  = [][]byte{keys1StakerKey, keys2StakerKey, keys3StakerKey, keys4StakerKey, keys5StakerKey}
)

func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func loadNodeID(stakeCert []byte) (string, error) {
	block, _ := pem.Decode(stakeCert)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("%w: problem parsing staking certificate", err)
	}

	id, err := ids.ToShortID(hashing.PubkeyBytesToAddress(cert.Raw))
	if err != nil {
		return "", fmt.Errorf("%w: problem deriving staker ID from certificate", err)
	}

	return id.PrefixedString(aConstants.NodeIDPrefix), nil
}

func NodeIDs() []string {
	nodeCerts := [][]byte{keys1StakerCrt, keys2StakerCrt, keys3StakerCrt, keys4StakerCrt, keys5StakerCrt}
	nodeIDs := make([]string, numNodes)
	for i, cert := range nodeCerts {
		id, err := loadNodeID(cert)
		if err != nil {
			panic(err)
		}
		nodeIDs[i] = id
	}
	return nodeIDs
}

func StartNetwork(ctx context.Context, configDir, vmPath string) error {
	dir, err := ioutil.TempDir("", "vm-tester")
	if err != nil {
		panic(err)
	}
	log.Println("created tmp dir", dir)

	// Copy files into custom plugins
	pluginsDir := fmt.Sprintf("%s/plugins", dir)
	if err := os.MkdirAll(pluginsDir, os.FileMode(0777)); err != nil {
		panic(err)
	}
	if err := Copy("system-plugins/evm", fmt.Sprintf("%s/evm", pluginsDir)); err != nil {
		panic(err)
	}
	if len(vmPath) > 0 {
		if err := Copy(vmPath, fmt.Sprintf("%s/%s", pluginsDir, constants.VMID)); err != nil {
			panic(err)
		}
	}

	nodeConfigs := make([]node.Config, numNodes)
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

		// TODO: create config directly instead of using flags
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
		if len(configDir) > 0 {
			df.ChainConfigDir = configDir
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
