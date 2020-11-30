package main

import (
	"fmt"
	"github.com/algorand/go-algorand-sdk/future"

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

	fromAddr := "C2MCKQYJCU4RNWCQVWWSWEHGPAD37BEMQTIMVD6XF36AUIPKOXWIZOO7ZE"
	toAddr := "H65XWDPZDEV7MXWDOUSLEL6UL6UO6CNK2ZILCENCFBKNCD4RATZIZZSIWQ"

	// Get the suggested transaction parameters
	txParams, err := algodClient.BuildSuggestedParams()
	if err != nil {
		fmt.Printf("error getting suggested tx params: %s\n", err)
		return
	}

	// how do we get this ?
	assetIndex := uint64(72303952)

	// Make transaction
	note := []byte("transfer to artist")
	//tx, err := future.MakePaymentTxn(fromAddr, toAddr, 1000, nil, "", txParams)
	tx, err := future.MakeAssetTransferTxn(fromAddr, toAddr, 1, note, txParams,"", assetIndex)
	if err != nil {
		fmt.Printf("Error creating transaction: %s\n", err)
		return
	}

	// Sign the transaction
	signResponse, err := kmdClient.SignTransaction(exampleWalletHandleToken, "password123", tx)
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

	//ubuntu@ns383889:~/uncopier$ go run transactAsset.go
	//Made a kmd client
	//Made an algod client
	//Got 2 wallet(s):
	//ID: 0377a9874ab7818c861ba465647e7e7b    Name: mylinuxwallet
	//found wallet 'mylinuxwallet' with ID: 0377a9874ab7818c861ba465647e7e7b
	//kmd made signed transaction with bytes: 82a3736967c440254141c22d4151a875cf5d1ff16ed9217987ca3a79163aec8405779bcbf7542445a558c671c783dd81c2125967126d5c742492f8ebb430cb1b2d333fe2c13405a374786e8ba461616d7401a461726376c4203fbb7b0df9192bf65ec37524b22fd45fa8ef09aad650b111a22854d10f9104f2a3666565cd03e8a26676ce009db645a367656eac6d61696e6e65742d76312e30a26768c420c061c4d8fc1dbdded2d7604be4568e3f6d041987ac37bde4b620b5ab39248adfa26c76ce009dba2da46e6f7465c4127472616e7366657220746f20617274697374a3736e64c4201698254309153916d850adad2b10e67807bf848c84d0ca8fd72efc0a21ea75eca474797065a56178666572a478616964ce044f4550
	//Transaction ID: 6E7LZUKJUP2JJSP73BJTNFIDP4NZKZTKTHEGBW4ZAJE7QSOLBYIQ
}
