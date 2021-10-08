package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/ava-labs/vm-tester/constants"
	"github.com/ava-labs/vm-tester/manager"
	"golang.org/x/sync/errgroup"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/api/info"
	"github.com/ava-labs/avalanchego/api/keystore"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

func main() {
	// Parse Args
	rawConfigDir := flag.String("config-dir", "", "directory for all VM configs")
	rawVMPath := flag.String("vm-path", "", "location of custom VM binary")
	rawVMGenesis := flag.String("vm-genesis", "", "location of custom VM genesis")
	flag.Parse()
	var configDir, vmPath, vmGenesis string
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
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return manager.StartNetwork(gctx, configDir, vmPath)
	})
	// only setup network if customVM exists
	if len(vmPath) > 0 {
		g.Go(func() error {
			return setupNetwork(gctx, vmGenesis)
		})
	}
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
	log.Fatal(g.Wait())
}

func setupNetwork(ctx context.Context, vmGenesis string) error {
	// wait for network to be bootstrapped
	// TODO: wait for all URLS to be good
	iClient := info.NewClient("http://localhost:9650", 10*time.Second)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		bootstrapped, _ := iClient.IsBootstrapped("P")
		if bootstrapped {
			break
		}

		fmt.Println("waiting for P-Chain to be bootstrapped")
		time.Sleep(1 * time.Second)
	}

	// create user
	userPass := api.UserPass{
		Username: "test",
		Password: "vmsrkewl",
	}
	kclient := keystore.NewClient("http://localhost:9650", 10*time.Second)
	ok, err := kclient.CreateUser(userPass)
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("could not create user")
	}

	// connect to local network
	client := platformvm.NewClient("http://localhost:9650", 10*time.Second)

	// Import genesis key
	fundedAddress, err := client.ImportKey(userPass, "PrivateKey-ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN")
	if err != nil {
		panic(err)
	}
	balance, err := client.GetBalance(fundedAddress)
	if err != nil {
		panic(err)
	}
	fmt.Println(fundedAddress, "Balance", balance)

	// create a subnet
	// TODO: migrate to not require pass on node
	subnetIDTx, err := client.CreateSubnet(userPass, []string{fundedAddress}, fundedAddress, []string{fundedAddress}, 1)
	if err != nil {
		panic(err)
	}

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		status, _ := client.GetTxStatus(subnetIDTx, true)
		if status.Status == platformvm.Committed {
			break
		}
		fmt.Println("waiting for subnet creation tx to be accepted", subnetIDTx)
		time.Sleep(1 * time.Second)
	}

	// get subnets (why don't we just get access?)
	subnets, err := client.GetSubnets([]ids.ID{})
	if err != nil {
		panic(err)
	}
	rSubnetID := subnets[0].ID
	subnetID := rSubnetID.String()

	// Add all validators to subnet (equal weight)
	for _, nodeID := range manager.NodeIDs() {
		txID, err := client.AddSubnetValidator(
			userPass,
			[]string{fundedAddress}, fundedAddress,
			subnetID, nodeID, 30,
			uint64(time.Now().Add(1*time.Minute).Unix()),
			uint64(time.Now().Add(30*24*time.Hour).Unix()),
		)
		if err != nil {
			panic(err)
		}

		for {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			status, _ := client.GetTxStatus(txID, true)
			if status.Status == platformvm.Committed {
				break
			}
			fmt.Println("waiting for add subnet validator tx to be accepted", txID)
			time.Sleep(1 * time.Second)
		}
	}

	// create blockchain
	genesis, err := ioutil.ReadFile(vmGenesis)
	if err != nil {
		panic(err)
	}
	txID, err := client.CreateBlockchain(
		userPass, []string{fundedAddress}, fundedAddress, rSubnetID,
		constants.VMID, []string{}, constants.VMName, genesis,
	)
	if err != nil {
		panic(err)
	}
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		status, _ := client.GetTxStatus(txID, true)
		if status.Status == platformvm.Committed {
			break
		}
		fmt.Println("waiting for create blockchain tx to be accepted", txID)
		time.Sleep(1 * time.Second)
	}

	// validate blockchain exists
	blockchains, err := client.GetBlockchains()
	if err != nil {
		panic(err)
	}
	var blockchainID ids.ID
	for _, blockchain := range blockchains {
		if blockchain.SubnetID == rSubnetID {
			blockchainID = blockchain.ID
			break
		}
	}
	if blockchainID == (ids.ID{}) {
		panic("could not find blockchain")
	}
	fmt.Println("blockchain created", blockchainID.String())

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		status, _ := client.GetBlockchainStatus(blockchainID.String())
		if status == platformvm.Validating {
			break
		}
		fmt.Println("waiting for validating status")
		time.Sleep(15 * time.Second)
	}
	fmt.Println("validating blockchain", blockchainID)
	return nil
}
