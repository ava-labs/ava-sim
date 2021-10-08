package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ava-labs/vm-tester/manager"
	"golang.org/x/sync/errgroup"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/api/info"
	"github.com/ava-labs/avalanchego/api/keystore"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

const (
	pluginDir          = "/Users/patrickogrady/code/avalanchego-internal/build/plugins"
	whitelistedSubnets = "p4jUwqZsA2LuSftroCd3zb4ytH8W99oXKuKVZdsty7eQ3rXD6"
)

func main() {
	// start local network
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return manager.StartNetwork(gctx, pluginDir, whitelistedSubnets)
	})
	g.Go(func() error {
		return setupNetwork(gctx)
	})
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

func setupNetwork(ctx context.Context) error {
	// wait for network to be bootstrapped
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

	// add validator to subnet
	nodeID, err := iClient.GetNodeID()
	if err != nil {
		panic(err)
	}
	// TODO: add all validators on subnet
	txID, err := client.AddSubnetValidator(
		userPass,
		[]string{fundedAddress}, fundedAddress,
		subnetID, nodeID, 30,
		uint64(time.Now().Add(5*time.Minute).Unix()),
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

	// create blockchain
	txID, err = client.CreateBlockchain(userPass, []string{fundedAddress}, fundedAddress, rSubnetID, "tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH", []string{}, "my vm", []byte("fP1vxkpyLWnH9dD6BQA"))
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
