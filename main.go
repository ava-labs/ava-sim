package main

import (
	// "bufio"
	"fmt"
	// "os"
	"time"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/api/info"
	"github.com/ava-labs/avalanchego/api/keystore"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

// pre-requisite
// * build timestampvm

func main() {
	// start local network
	// runscript scripts/five_node_staking.lua (modify gossip frequency)

	// wait for network to be bootstrapped
	iClient := info.NewClient("http://localhost:9650", 10*time.Second)
	for {
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
		status, _ := client.GetTxStatus(txID, true)
		if status.Status == platformvm.Committed {
			break
		}
		fmt.Println("waiting for add subnet validator tx to be accepted", txID)
		time.Sleep(1 * time.Second)
	}

	// whitelist subnet (do from API?)
	// fmt.Println("whitelist", subnetID)
	// fmt.Print("Press 'Enter' to continue...")
	// bufio.NewReader(os.Stdin).ReadBytes('\n')
	// Always: p4jUwqZsA2LuSftroCd3zb4ytH8W99oXKuKVZdsty7eQ3rXD6

	// create genesis
	// hardcoded: fP1vxkpyLWnH9dD6BQA ("helloworld")

	// create blockchain
	txID, err = client.CreateBlockchain(userPass, []string{fundedAddress}, fundedAddress, rSubnetID, "tGas3T58KzdjLHhBDMnH2TvrddhqTji5iZAMZ3RXs2NLpSnhH", []string{}, "my vm", []byte("fP1vxkpyLWnH9dD6BQA"))
	if err != nil {
		panic(err)
	}
	for {
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
		status, _ := client.GetBlockchainStatus(blockchainID.String())
		if status == platformvm.Validating {
			break
		}
		fmt.Println("waiting for validating status, got", status)
		time.Sleep(15 * time.Second)
	}
	fmt.Println("validating blockchain", blockchainID)

	// TODO: interact with blockchain
}
