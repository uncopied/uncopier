package main

import (
	"fmt"
	"github.com/algorand/go-algorand-sdk/future"

	"github.com/algorand/go-algorand-sdk/client/algod"
	"github.com/algorand/go-algorand-sdk/client/kmd"
	"github.com/algorand/go-algorand-sdk/transaction"
)

// CHANGE ME
const kmdAddress = "http://ns383889.ip-46-105-120.eu:7833"
const kmdToken = "206ba3f9ad1d83523fb2a303dd055cd99ce10c5be01e35ee88285fe51438f02a"
const algodAddress = "http://ns383889.ip-46-105-120.eu:8080"
const algodToken = "3005189549ea62dc24fac6102493c941f68840304d3819727b83d3eb74de8e6d"

func main() {
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

	fromAddr := "C2MCKQYJCU4RNWCQVWWSWEHGPAD37BEMQTIMVD6XF36AUIPKOXWIZOO7ZE"
	toAddr := "H65XWDPZDEV7MXWDOUSLEL6UL6UO6CNK2ZILCENCFBKNCD4RATZIZZSIWQ"

	// Get the suggested transaction parameters
	txParams, err := algodClient.SuggestedParams()
	if err != nil {
		fmt.Printf("error getting suggested tx params: %s\n", err)
		return
	}

	// Make transaction
	tx, err := future.MakePaymentTxn(fromAddr, toAddr, 1000, nil, "", txParams)
	if err != nil {
		fmt.Printf("Error creating transaction: %s\n", err)
		return
	}

	// Sign the transaction
	signResponse, err := kmdClient.SignTransaction(exampleWalletHandleToken, "testpassword", tx)
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
