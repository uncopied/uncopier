package main

import (
	"fmt"
	transaction "github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/client/algod"
	"github.com/algorand/go-algorand-sdk/client/kmd"
)


func main() {
	const kmdAddress = "http://localhost:7833"
	const kmdToken = "206ba3f9ad1d83523fb2a303dd055cd99ce10c5be01e35ee88285fe51438f02a"
	const algodAddress = "http://localhost:8080"
	const algodToken = "da61ace80780af7b1c78456c7d1d2a511758a754d2c219e1a6b37c32763f5bfe"

	// Create a kmd client
	kmdClient, err := kmd.MakeClient(kmdAddress, kmdToken)
	if err != nil {
		fmt.Printf("failed to make kmd client: %s\n", err)
		return
	}
	fmt.Println("Made a kmd client")

	// Create an algod client
	algodClient, err := algod.MakeClient(algodAddress, algodToken)
	if err != nil {
		fmt.Printf("failed to make algod client: %s\n", err)
		return
	}
	fmt.Println("Made an algod client")
	// Get the list of wallets
	listResponse, err := kmdClient.ListWallets()
	if err != nil {
		fmt.Printf("error listing wallets: %s\n", err)
		return
	}


	// Find our wallet name in the list
	var exampleWalletID string
	fmt.Printf("Got %d wallet(s):\n", len(listResponse.Wallets))
	for _, wallet := range listResponse.Wallets {
		fmt.Printf("ID: %s\tName: %s\n", wallet.ID, wallet.Name)
		if wallet.Name == "mylinuxwallet" {
			fmt.Printf("found wallet '%s' with ID: %s\n", wallet.Name, wallet.ID)
			exampleWalletID = wallet.ID
			break
		}
	}

	// Get a wallet handle. The wallet handle is used for things like signing transactions
	// and creating accounts. Wallet handles do expire, but they can be renewed
	initResponse, err := kmdClient.InitWalletHandle(exampleWalletID, "password123")
	if err != nil {
		fmt.Printf("Error initializing wallet handle: %s\n", err)
		return
	}

	// Extract the wallet handle
	exampleWalletHandleToken := initResponse.WalletHandleToken
	addr := "C2MCKQYJCU4RNWCQVWWSWEHGPAD37BEMQTIMVD6XF36AUIPKOXWIZOO7ZE"

	defaultFrozen := false // whether user accounts will need to be unfrozen before transacting
	totalIssuance := uint64(1) // total number of this asset in circulation
	decimals := uint32(0) // hint to GUIs for interpreting base unit
	reserve := addr // specified address is considered the asset reserve (it has no special privileges, this is only informational)
	freeze := addr // specified address can freeze or unfreeze user asset holdings
	clawback := addr // specified address can revoke user asset holdings and send them to other addresses
	manager := addr // specified address can change reserve, freeze, clawback, and manager
	unitName := "lady-1/15" // used to display asset units to user
	assetName := "Portrait of a Lady (1/15)" // "friendly name" of asset
	note := []byte("test asset create") // arbitrary data to be stored in the transaction; here, none is stored
	assetURL := "https://uncopied.org/xEkMZ.png" // optional string pointing to a URL relating to the asset. 32 character length.
	assetMetadataHash := "0bc777329a2919fd6ffa96bace9a4779" // optional hash commitment of some sort relating to the asset. 32 character length.

	// Get the suggested transaction parameters
	txParams, err := algodClient.BuildSuggestedParams()
	if err != nil {
		fmt.Printf("error getting suggested tx params: %s\n", err)
		return
	}

	// signing and sending "txn" allows "addr" to create an asset
	txn, err := transaction.MakeAssetCreateTxn(addr, note, txParams,
		totalIssuance, decimals, defaultFrozen, manager, reserve, freeze, clawback,
		unitName, assetName, assetURL, assetMetadataHash)
	if err != nil {
		fmt.Printf("Failed to make asset: %s\n", err)
		return
	}
	fmt.Printf("Asset created AssetName: %s\n", txn.AssetConfigTxnFields.AssetParams.AssetName)
	// Sign the transaction
	signResponse, err := kmdClient.SignTransaction(exampleWalletHandleToken, "password123", txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction with kmd: %s\n", err)
		return
	}

	fmt.Printf("kmd made signed transaction with bytes: %x\n", signResponse.SignedTransaction)

	// Broadcast the transaction to the network
	// Note that this transaction will get rejected because the accounts do not have any tokens
	sendResponse, err := algodClient.SendRawTransaction(signResponse.SignedTransaction)
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return
	}

	fmt.Printf("Transaction ID: %s\n", sendResponse.TxID)
}
