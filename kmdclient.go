package main

import (
	"fmt"

	"github.com/algorand/go-algorand-sdk/client/kmd"
	"github.com/algorand/go-algorand-sdk/types"
)

// These constants represent the kmdd REST endpoint and the corresponding API
// token. You can retrieve these from the `kmd.net` and `kmd.token` files in
// the kmd data directory.
const kmdAddress = "http://localhost:7833"
const kmdToken = "206ba3f9ad1d83523fb2a303dd055cd99ce10c5be01e35ee88285fe51438f02a"

func main() {
	// Create a kmd client
	kmdClient, err := kmd.MakeClient(kmdAddress, kmdToken)
	if err != nil {
		fmt.Printf("failed to make kmd client: %s\n", err)
		return
	}
	fmt.Println("Made a kmd client")

	// Create the example wallet, if it doesn't already exist
	cwResponse, err := kmdClient.CreateWallet("mylinuxwallet", "password123", kmd.DefaultWalletDriver, types.MasterDerivationKey{})
	if err != nil {
		fmt.Printf("error creating wallet: %s\n", err)
		return
	}

	// We need the wallet ID in order to get a wallet handle, so we can add accounts
	exampleWalletID := cwResponse.Wallet.ID
	fmt.Printf("Created wallet '%s' with ID: %s\n", cwResponse.Wallet.Name, exampleWalletID)

	// Get a wallet handle. The wallet handle is used for things like signing transactions
	// and creating accounts. Wallet handles do expire, but they can be renewed
	initResponse, err := kmdClient.InitWalletHandle(exampleWalletID, "password123")
	if err != nil {
		fmt.Printf("Error initializing wallet handle: %s\n", err)
		return
	}

	// Extract the wallet handle
	exampleWalletHandleToken := initResponse.WalletHandleToken

	// Generate a new address from the wallet handle
	genResponse, err := kmdClient.GenerateKey(exampleWalletHandleToken)
	if err != nil {
		fmt.Printf("Error generating key: %s\n", err)
		return
	}
	fmt.Printf("Generated address %s\n", genResponse.Address)
}
