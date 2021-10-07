package manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/ava-labs/avalanchego/app/process"
	"github.com/ava-labs/avalanchego/config"
	"github.com/ava-labs/avalanchego/node"
	"github.com/hashicorp/go.net/context"
	"golang.org/x/sync/errgroup"
)

const (
	bootstrapID  = "NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg"
	bootstrapIP  = "localhost:9651"
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
	// TODO: add path to plugin to copy to plugin dirs
	// TODO add name
	dir, err := ioutil.TempDir("vm-tester", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	nodeConfigs := make([]node.Config, numNodes)
	nodeCerts := [][]byte{keys1StakerCrt, keys2StakerCrt, keys3StakerCrt, keys4StakerCrt, keys4StakerCrt}
	nodeKeys := [][]byte{keys1StakerKey, keys2StakerKey, keys3StakerKey, keys4StakerKey, keys5StakerKey}
	for i := 0; i < numNodes; i++ {
		nodeDir := fmt.Sprintf("%s/node%d", dir, i+1)
		if err := os.Mkdir(nodeDir, os.FileMode(644)); err != nil {
			panic(err)
		}
		certFile := fmt.Sprintf("%s/staker.crt", nodeDir)
		if err := ioutil.WriteFile(certFile, nodeCerts[i], os.FileMode(644)); err != nil {
			panic(err)
		}
		keyFile := fmt.Sprintf("%s/staker.key", nodeDir)
		if err := ioutil.WriteFile(keyFile, nodeKeys[i], os.FileMode(644)); err != nil {
			panic(err)
		}

		df := defaultFlags()
		df.LogLevel = "info"
		df.StakingEnabled = true
		df.HTTPPort = uint(baseHTTPPort + i)
		df.StakingPort = uint(baseHTTPPort + i + 1)
		df.BootstrapIPs = bootstrapIP
		df.BootstrapIDs = bootstrapID
		df.WhitelistedSubnets = "TODO"
		df.StakingTLSCertFile = certFile
		df.StakingTLSKeyFile = keyFile
		df.PluginDir = pluginDir
		df.WhitelistedSubnets = whitelistedSubnets
		nodeConfig, err := createNodeConfig(pluginDir, flagsToArgs(df))
		if err != nil {
			panic(err)
		}
		nodeConfigs[i] = nodeConfig
	}

	// register signals to kill the application
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	signal.Notify(signals, syscall.SIGTERM)
	defer func() {
		// shut down the signal go routine
		signal.Stop(signals)
		close(signals)
	}

	g, gctx := errgroup.WithContext(ctx)
	ggctx, cancel := context.WithCancel(gctx)
	g.Go(func() error {
		select {
		case <-signal:
			cancel()
		case <-ggctx.Done():
		}
		return nil
	})
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

	// start running the application
	if err := app.Start(); err != nil {
		return err
	}

	g.Go(func() error {
		<-ctx.Done()
		return app.Stop()
	})

	exitCode, err := app.ExitCode()
	fmt.Fprintf(os.Stderr, "exit code %v\n", exitCode)
	return err
}

func createNodeConfig(pluginDir string, args []string) (node.Config, error) {
	fs := config.BuildFlagSet()
	v, err := config.BuildViper(fs, args)
	if err != nil {
		return node.Config{}, err
	}

	return config.GetNodeConfig(v, pluginDir)
}

// Flags represents available CLI flags when starting a node
type Flags struct {
	// Assertions
	AssertionsEnabled bool

	// Version
	Version bool

	// TX fees
	TxFee uint

	// IP
	PublicIP              string
	DynamicUpdateDuration string
	DynamicPublicIP       string

	// Network ID
	NetworkID string

	// Crypto
	SignatureVerificationEnabled bool

	// APIs
	APIAdminEnabled    bool
	APIIPCsEnabled     bool
	APIKeystoreEnabled bool
	APIMetricsEnabled  bool
	APIHealthEnabled   bool
	APIInfoEnabled     bool

	// HTTP
	HTTPHost        string
	HTTPPort        uint
	HTTPTLSEnabled  bool
	HTTPTLSCertFile string
	HTTPTLSKeyFile  string

	// Bootstrapping
	BootstrapIPs                     string
	BootstrapIDs                     string
	BootstrapBeaconConnectionTimeout string

	// Build
	BuildDir string

	// Plugins
	PluginDir string

	// Logging
	LogLevel            string
	LogDir              string
	LogDisplayLevel     string
	LogDisplayHighlight string

	// Consensus
	SnowAvalancheBatchSize      int
	SnowAvalancheNumParents     int
	SnowSampleSize              int
	SnowQuorumSize              int
	SnowVirtuousCommitThreshold int
	SnowRogueCommitThreshold    int
	SnowEpochFirstTransition    int
	SnowEpochDuration           string
	SnowConcurrentRepolls       int
	MinDelegatorStake           int
	ConsensusShutdownTimeout    string
	ConsensusGossipFrequency    string
	MinDelegationFee            int
	MinValidatorStake           int
	MaxStakeDuration            string
	MaxValidatorStake           int

	// Staking
	StakingEnabled        bool
	StakeMintingPeriod    string
	StakingPort           uint
	StakingDisabledWeight int
	StakingTLSKeyFile     string
	StakingTLSCertFile    string

	// Auth
	APIAuthRequired        bool
	APIAuthPasswordFileKey string
	MinStakeDuration       string

	// Whitelisted Subnets
	WhitelistedSubnets string

	// Config
	ConfigFile string

	// IPCS
	IPCSChainIDs string
	IPCSPath     string

	// File Descriptor Limit
	FDLimit int

	// Benchlist
	BenchlistFailThreshold      int
	BenchlistMinFailingDuration string
	BenchlistPeerSummaryEnabled bool
	BenchlistDuration           string
	// Network Timeout
	NetworkInitialTimeout                   string
	NetworkMinimumTimeout                   string
	NetworkMaximumTimeout                   string
	NetworkHealthMaxSendFailRateKey         float64
	NetworkHealthMaxPortionSendQueueFillKey float64
	NetworkHealthMaxTimeSinceMsgSentKey     string
	NetworkHealthMaxTimeSinceMsgReceivedKey string
	NetworkHealthMinConnPeers               int
	NetworkTimeoutCoefficient               int
	NetworkTimeoutHalflife                  string

	// Peer List Gossiping
	NetworkPeerListGossipFrequency string
	NetworkPeerListGossipSize      int
	NetworkPeerListSize            int

	// Uptime Requirement
	UptimeRequirement float64

	// Retry
	RetryBootstrap bool

	// Health
	HealthCheckAveragerHalflifeKey string
	HealthCheckFreqKey             string

	// Router
	RouterHealthMaxOutstandingRequestsKey int
	RouterHealthMaxDropRateKey            float64

	IndexEnabled bool

	PluginModeEnabled bool
}

// defaultFlags returns Avash-specific default node flags
func defaultFlags() Flags {
	return Flags{
		AssertionsEnabled:                       true,
		Version:                                 false,
		TxFee:                                   1000000,
		PublicIP:                                "127.0.0.1",
		DynamicUpdateDuration:                   "5m",
		DynamicPublicIP:                         "",
		NetworkID:                               "local",
		SignatureVerificationEnabled:            true,
		APIAdminEnabled:                         true,
		APIIPCsEnabled:                          true,
		APIKeystoreEnabled:                      true,
		APIMetricsEnabled:                       true,
		HTTPHost:                                "127.0.0.1",
		HTTPPort:                                9650,
		HTTPTLSEnabled:                          false,
		HTTPTLSCertFile:                         "",
		HTTPTLSKeyFile:                          "",
		BootstrapIPs:                            "",
		BootstrapIDs:                            "",
		BootstrapBeaconConnectionTimeout:        "60s",
		BuildDir:                                "",
		PluginDir:                               "",
		LogLevel:                                "info",
		LogDir:                                  baseLogs,
		LogDisplayLevel:                         "", // defaults to the value provided to --log-level
		LogDisplayHighlight:                     "colors",
		SnowAvalancheBatchSize:                  30,
		SnowAvalancheNumParents:                 5,
		SnowSampleSize:                          2,
		SnowQuorumSize:                          2,
		SnowVirtuousCommitThreshold:             5,
		SnowRogueCommitThreshold:                10,
		SnowEpochFirstTransition:                1609873200,
		SnowEpochDuration:                       "6h",
		SnowConcurrentRepolls:                   4,
		MinDelegatorStake:                       5000000,
		ConsensusShutdownTimeout:                "5s",
		ConsensusGossipFrequency:                "10s",
		MinDelegationFee:                        20000,
		MinValidatorStake:                       5000000,
		MaxStakeDuration:                        "8760h",
		MaxValidatorStake:                       3000000000000000,
		StakeMintingPeriod:                      "8760h",
		NetworkInitialTimeout:                   "5s",
		NetworkMinimumTimeout:                   "5s",
		NetworkMaximumTimeout:                   "10s",
		NetworkHealthMaxSendFailRateKey:         0.9,
		NetworkHealthMaxPortionSendQueueFillKey: 0.9,
		NetworkHealthMaxTimeSinceMsgSentKey:     "1m",
		NetworkHealthMaxTimeSinceMsgReceivedKey: "1m",
		NetworkHealthMinConnPeers:               1,
		NetworkTimeoutCoefficient:               2,
		NetworkTimeoutHalflife:                  "5m",
		NetworkPeerListGossipFrequency:          "1m",
		NetworkPeerListGossipSize:               50,
		NetworkPeerListSize:                     20,
		StakingEnabled:                          false,
		StakingPort:                             9651,
		StakingDisabledWeight:                   1,
		StakingTLSKeyFile:                       "",
		StakingTLSCertFile:                      "",
		APIAuthRequired:                         false,
		APIAuthPasswordFileKey:                  "",
		MinStakeDuration:                        "336h",
		APIHealthEnabled:                        true,
		ConfigFile:                              "",
		WhitelistedSubnets:                      "",
		APIInfoEnabled:                          true,
		IPCSChainIDs:                            "",
		IPCSPath:                                "/tmp",
		FDLimit:                                 32768,
		BenchlistDuration:                       "1h",
		BenchlistFailThreshold:                  10,
		BenchlistMinFailingDuration:             "5m",
		BenchlistPeerSummaryEnabled:             false,
		UptimeRequirement:                       0.6,
		RetryBootstrap:                          true,
		HealthCheckAveragerHalflifeKey:          "10s",
		HealthCheckFreqKey:                      "30s",
		RouterHealthMaxOutstandingRequestsKey:   1024,
		RouterHealthMaxDropRateKey:              1,
		IndexEnabled:                            false,
		PluginModeEnabled:                       false,
	}
}

// flagsToArgs converts a `Flags` struct into a CLI command flag string
func flagsToArgs(flags Flags) []string {
	// Port targets
	httpPortString := strconv.FormatUint(uint64(flags.HTTPPort), 10)
	stakingPortString := strconv.FormatUint(uint64(flags.StakingPort), 10)

	// Paths/directories
	dbPath := baseDB + "/" + stakingPortString
	logPath := baseLogs + "/" + stakingPortString

	wd, _ := os.Getwd()
	// If the path given in the flag doesn't begin with "/", treat it as relative
	// to the directory of the avash binary
	httpCertFile := flags.HTTPTLSCertFile
	if httpCertFile != "" && string(httpCertFile[0]) != "/" {
		httpCertFile = fmt.Sprintf("%s/%s", wd, httpCertFile)
	}

	httpKeyFile := flags.HTTPTLSKeyFile
	if httpKeyFile != "" && string(httpKeyFile[0]) != "/" {
		httpKeyFile = fmt.Sprintf("%s/%s", wd, httpKeyFile)
	}

	stakerCertFile := flags.StakingTLSCertFile
	if stakerCertFile != "" && string(stakerCertFile[0]) != "/" {
		stakerCertFile = fmt.Sprintf("%s/%s", wd, stakerCertFile)
	}

	stakerKeyFile := flags.StakingTLSKeyFile
	if stakerKeyFile != "" && string(stakerKeyFile[0]) != "/" {
		stakerKeyFile = fmt.Sprintf("%s/%s", wd, stakerKeyFile)
	}

	args := []string{
		"--assertions-enabled=" + strconv.FormatBool(flags.AssertionsEnabled),
		"--version=" + strconv.FormatBool(flags.Version),
		"--tx-fee=" + strconv.FormatUint(uint64(flags.TxFee), 10),
		"--public-ip=" + flags.PublicIP,
		"--dynamic-update-duration=" + flags.DynamicUpdateDuration,
		"--dynamic-public-ip=" + flags.DynamicPublicIP,
		"--network-id=" + flags.NetworkID,
		"--signature-verification-enabled=" + strconv.FormatBool(flags.SignatureVerificationEnabled),
		"--api-admin-enabled=" + strconv.FormatBool(flags.APIAdminEnabled),
		"--api-ipcs-enabled=" + strconv.FormatBool(flags.APIIPCsEnabled),
		"--api-keystore-enabled=" + strconv.FormatBool(flags.APIKeystoreEnabled),
		"--api-metrics-enabled=" + strconv.FormatBool(flags.APIMetricsEnabled),
		"--http-host=" + flags.HTTPHost,
		"--http-port=" + httpPortString,
		"--http-tls-enabled=" + strconv.FormatBool(flags.HTTPTLSEnabled),
		"--http-tls-cert-file=" + httpCertFile,
		"--http-tls-key-file=" + httpKeyFile,
		"--bootstrap-ips=" + flags.BootstrapIPs,
		"--bootstrap-ids=" + flags.BootstrapIDs,
		"--bootstrap-beacon-connection-timeout=" + flags.BootstrapBeaconConnectionTimeout,
		"--db-dir=" + dbPath,
		// "--db-type=memdb",
		"--plugin-dir=" + flags.PluginDir,
		"--build-dir=" + flags.BuildDir,
		"--log-level=" + flags.LogLevel,
		"--log-dir=" + logPath,
		"--log-display-level=" + flags.LogDisplayLevel,
		"--log-display-highlight=" + flags.LogDisplayHighlight,
		"--snow-avalanche-batch-size=" + strconv.Itoa(flags.SnowAvalancheBatchSize),
		"--snow-avalanche-num-parents=" + strconv.Itoa(flags.SnowAvalancheNumParents),
		"--snow-sample-size=" + strconv.Itoa(flags.SnowSampleSize),
		"--snow-quorum-size=" + strconv.Itoa(flags.SnowQuorumSize),
		"--snow-virtuous-commit-threshold=" + strconv.Itoa(flags.SnowVirtuousCommitThreshold),
		"--snow-rogue-commit-threshold=" + strconv.Itoa(flags.SnowRogueCommitThreshold),
		"--snow-epoch-first-transition=" + strconv.Itoa(flags.SnowEpochFirstTransition),
		"--snow-epoch-duration=" + flags.SnowEpochDuration,
		"--min-delegator-stake=" + strconv.Itoa(flags.MinDelegatorStake),
		"--consensus-shutdown-timeout=" + flags.ConsensusShutdownTimeout,
		"--consensus-gossip-frequency=" + flags.ConsensusGossipFrequency,
		"--min-delegation-fee=" + strconv.Itoa(flags.MinDelegationFee),
		"--min-validator-stake=" + strconv.Itoa(flags.MinValidatorStake),
		"--max-stake-duration=" + flags.MaxStakeDuration,
		"--max-validator-stake=" + strconv.Itoa(flags.MaxValidatorStake),
		"--snow-concurrent-repolls=" + strconv.Itoa(flags.SnowConcurrentRepolls),
		"--stake-minting-period=" + flags.StakeMintingPeriod,
		"--network-initial-timeout=" + flags.NetworkInitialTimeout,
		"--network-minimum-timeout=" + flags.NetworkMinimumTimeout,
		"--network-maximum-timeout=" + flags.NetworkMaximumTimeout,
		fmt.Sprintf("--network-health-max-send-fail-rate=%f", flags.NetworkHealthMaxSendFailRateKey),
		fmt.Sprintf("--network-health-max-portion-send-queue-full=%f", flags.NetworkHealthMaxPortionSendQueueFillKey),
		"--network-health-max-time-since-msg-sent=" + flags.NetworkHealthMaxTimeSinceMsgSentKey,
		"--network-health-max-time-since-msg-received=" + flags.NetworkHealthMaxTimeSinceMsgReceivedKey,
		"--network-health-min-conn-peers=" + strconv.Itoa(flags.NetworkHealthMinConnPeers),
		"--network-timeout-coefficient=" + strconv.Itoa(flags.NetworkTimeoutCoefficient),
		"--network-timeout-halflife=" + flags.NetworkTimeoutHalflife,
		"--network-peer-list-gossip-frequency=" + flags.NetworkPeerListGossipFrequency,
		"--network-peer-list-gossip-size=" + strconv.Itoa(flags.NetworkPeerListGossipSize),
		"--network-peer-list-size=" + strconv.Itoa(flags.NetworkPeerListSize),
		"--staking-enabled=" + strconv.FormatBool(flags.StakingEnabled),
		"--staking-port=" + stakingPortString,
		"--staking-disabled-weight=" + strconv.Itoa(flags.StakingDisabledWeight),
		"--staking-tls-key-file=" + stakerKeyFile,
		"--staking-tls-cert-file=" + stakerCertFile,
		"--api-auth-required=" + strconv.FormatBool(flags.APIAuthRequired),
		"--api-auth-password-file=" + flags.APIAuthPasswordFileKey,
		"--min-stake-duration=" + flags.MinStakeDuration,
		"--whitelisted-subnets=" + flags.WhitelistedSubnets,
		"--api-health-enabled=" + strconv.FormatBool(flags.APIHealthEnabled),
		"--config-file=" + flags.ConfigFile,
		"--api-info-enabled=" + strconv.FormatBool(flags.APIInfoEnabled),
		"--ipcs-chain-ids=" + flags.IPCSChainIDs,
		"--ipcs-path=" + flags.IPCSPath,
		"--fd-limit=" + strconv.Itoa(flags.FDLimit),
		"--benchlist-duration=" + flags.BenchlistDuration,
		"--benchlist-fail-threshold=" + strconv.Itoa(flags.BenchlistFailThreshold),
		"--benchlist-min-failing-duration=" + flags.BenchlistMinFailingDuration,
		"--benchlist-peer-summary-enabled=" + strconv.FormatBool(flags.BenchlistPeerSummaryEnabled),
		fmt.Sprintf("--uptime-requirement=%f", flags.UptimeRequirement),
		"--bootstrap-retry-enabled=" + strconv.FormatBool(flags.RetryBootstrap),
		"--health-check-averager-halflife=" + flags.HealthCheckAveragerHalflifeKey,
		"--health-check-frequency=" + flags.HealthCheckFreqKey,
		"--router-health-max-outstanding-requests=" + strconv.Itoa(flags.RouterHealthMaxOutstandingRequestsKey),
		fmt.Sprintf("--router-health-max-drop-rate=%f", flags.RouterHealthMaxDropRateKey),
		"--index-enabled=" + strconv.FormatBool(flags.IndexEnabled),
		"--plugin-mode-enabled=" + strconv.FormatBool(flags.PluginModeEnabled),
		// "--coreth-config=" + "{\"eth-api-enabled\": true, \"rpc-gas-cap\": 2500000000, \"debug-api-enabled\": true,  \"rpc-tx-fee-cap\": 100, \"tx-pool-api-enabled\": true, \"pruning-enabled\": false}",
	}
	args = removeEmptyFlags(args)

	return args
}

func removeEmptyFlags(args []string) []string {
	var res []string
	for _, f := range args {
		tmp := strings.TrimSpace(f)
		if !strings.HasSuffix(tmp, "=") {
			res = append(res, tmp)
		}
	}
	return res
}
