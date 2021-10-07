package main

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/api/info"
	"github.com/ava-labs/avalanchego/api/keystore"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

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

	// Import genesis key
	userPass := api.UserPass{
		Username: "test",
		Password: "vmsrkewl",
	}

	// create user
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

	fundedAddress, err := client.ImportKey(userPass, "PrivateKey-vmRQiZeXEXYMyJhEiqdC2z5JhuDbxL8ix9UVvjgMu2Er1NepE")
	if err != nil {
		panic(err)
	}
	balance, err := client.GetBalance(fundedAddress)
	if err != nil {
		panic(err)
	}
	fmt.Println(fundedAddress, "Balance", balance)

	fundedAddress, err = client.ImportKey(userPass, "PrivateKey-ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN")
	if err != nil {
		panic(err)
	}
	balance, err = client.GetBalance(fundedAddress)
	if err != nil {
		panic(err)
	}
	fmt.Println(fundedAddress, "Balance", balance)

	// create a subnet

	// add validator to subnet

	// whitelist subnet (do from API?)

	// create genesis

	// create blockchain

	// validate blockchain exists

	// interact with blockchain
	fmt.Println("vim-go")
}
