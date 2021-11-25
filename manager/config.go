package manager

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ava-labs/avalanchego/config"
	"github.com/ava-labs/avalanchego/node"
)

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

	// DB
	DBDir string

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
	ConfigFile     string
	ChainConfigDir string

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
		LogLevel:                                "info",
		LogDisplayLevel:                         "", // defaults to the value provided to --log-level
		LogDisplayHighlight:                     "colors",
		SnowAvalancheBatchSize:                  30,
		SnowAvalancheNumParents:                 5,
		SnowSampleSize:                          2,
		SnowQuorumSize:                          2,
		SnowVirtuousCommitThreshold:             5,
		SnowRogueCommitThreshold:                10,
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
		NetworkPeerListGossipFrequency:          "1s",
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
	fmt.Println(stakerKeyFile, stakerCertFile)

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
		"--db-dir=" + flags.DBDir,
		"--build-dir=" + flags.BuildDir,
		"--log-level=" + flags.LogLevel,
		"--log-dir=" + flags.LogDir,
		"--log-display-level=" + flags.LogDisplayLevel,
		"--log-display-highlight=" + flags.LogDisplayHighlight,
		"--snow-avalanche-batch-size=" + strconv.Itoa(flags.SnowAvalancheBatchSize),
		"--snow-avalanche-num-parents=" + strconv.Itoa(flags.SnowAvalancheNumParents),
		"--snow-sample-size=" + strconv.Itoa(flags.SnowSampleSize),
		"--snow-quorum-size=" + strconv.Itoa(flags.SnowQuorumSize),
		"--snow-virtuous-commit-threshold=" + strconv.Itoa(flags.SnowVirtuousCommitThreshold),
		"--snow-rogue-commit-threshold=" + strconv.Itoa(flags.SnowRogueCommitThreshold),
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
		"--chain-config-dir=" + flags.ChainConfigDir,
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
