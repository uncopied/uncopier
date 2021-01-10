package blockchain

import (
	"fmt"
	"github.com/algorand/go-algorand-sdk/client/algod"
	"github.com/algorand/go-algorand-sdk/client/kmd"
	transaction "github.com/algorand/go-algorand-sdk/future"
	"github.com/google/uuid"
	"github.com/uncopied/uncopier/database/dbmodel"
	"os"
)

func AlgorandCreateNFT(asset *dbmodel.Asset, assetViewURL string, metadataHash string) (string, error) {
	createAlgorandAsset := os.Getenv("ALGORAND_CREATE_ASSET")
	if createAlgorandAsset == "true" {
		return AlgorandCreateNFT_(asset, assetViewURL, metadataHash)
	} else {
		return uuid.New().String(),nil
	}
}

func AlgorandCreateNFT_(asset *dbmodel.Asset, assetViewURL string, metadataHash string) (string, error) {

	kmdAddress := os.Getenv("ALGORAND_KMDADDRESS")         // "http://localhost:7833"
	kmdToken := os.Getenv("ALGORAND_KMDTOKEN")             //"206ba3f9ad1d83523fb2a303dd055cd99ce10c5be01e35ee88285fe51438f02a"
	algodAddress := os.Getenv("ALGORAND_ALGODADDRESS")     //"http://localhost:8080"
	algodToken := os.Getenv("ALGORAND_ALGODTOKEN")         // "da61ace80780af7b1c78456c7d1d2a511758a754d2c219e1a6b37c32763f5bfe"
	walletName := os.Getenv("ALGORAND_WALLETNAME")         //"mylinuxwallet"
	walletPassword := os.Getenv("ALGORAND_WALLETPASSWORD") // "password123"
	accountAddr := os.Getenv("ALGORAND_ACCOUNTADDR")       //"C2MCKQYJCU4RNWCQVWWSWEHGPAD37BEMQTIMVD6XF36AUIPKOXWIZOO7ZE"

	// Create a kmd client
	kmdClient, err := kmd.MakeClient(kmdAddress, kmdToken)
	if err != nil {
		fmt.Printf("failed to make kmd client: %s\n", err)
		return "",err
	}
	fmt.Println("Made a kmd client")

	// Create an algod client
	algodClient, err := algod.MakeClient(algodAddress, algodToken)
	if err != nil {
		fmt.Printf("failed to make algod client: %s\n", err)
		return "",err
	}
	fmt.Println("Made an algod client")
	// Get the list of wallets
	listResponse, err := kmdClient.ListWallets()
	if err != nil {
		fmt.Printf("error listing wallets: %s\n", err)
		return "",err
	}


	// Find our wallet name in the list
	var exampleWalletID string
	fmt.Printf("Got %d wallet(s):\n", len(listResponse.Wallets))
	for _, wallet := range listResponse.Wallets {
		fmt.Printf("ID: %s\tName: %s\n", wallet.ID, wallet.Name)
		if wallet.Name == walletName {
			fmt.Printf("found wallet '%s' with ID: %s\n", wallet.Name, wallet.ID)
			exampleWalletID = wallet.ID
			break
		}
	}

	// Get a wallet handle. The wallet handle is used for things like signing transactions
	// and creating accounts. Wallet handles do expire, but they can be renewed
	initResponse, err := kmdClient.InitWalletHandle(exampleWalletID, walletPassword)
	if err != nil {
		fmt.Printf("Error initializing wallet handle: %s\n", err)
		return "",err
	}

	// Extract the wallet handle
	exampleWalletHandleToken := initResponse.WalletHandleToken

	defaultFrozen := false              // whether user accounts will need to be unfrozen before transacting
	totalIssuance := uint64(1)          // total number of this asset in circulation
	decimals := uint32(0)               // hint to GUIs for interpreting base unit
	reserve := accountAddr              // specified address is considered the asset reserve (it has no special privileges, this is only informational)
	freeze := accountAddr               // specified address can freeze or unfreeze user asset holdings
	clawback := accountAddr             // specified address can revoke user asset holdings and send them to other addresses
	manager := accountAddr              // specified address can change reserve, freeze, clawback, and manager
	unitName := ""                      // used to display asset units to user
	assetName := asset.AssetLabel       // "friendly name" of asset
	noteStr := asset.Note
	if noteStr == "" {
		noteStr = asset.CertificateLabel
	}
	note := []byte(noteStr)          // arbitrary data to be stored in the transaction; here, none is stored

	//assetURL := "https://uncopied.org/c/v/"+strconv.Itoa(int(asset.ID)) // optional string pointing to a URL relating to the asset. 32 character length.
	assetMetadataHash := metadataHash // optional hash commitment of some sort relating to the asset. 32 character length.

	// Get the suggested transaction parameters
	txParams, err := algodClient.BuildSuggestedParams()
	if err != nil {
		fmt.Printf("error getting suggested tx params: %s\n", err)
		return "",err
	}

	// signing and sending "txn" allows "accountAddr" to create an asset
	txn, err := transaction.MakeAssetCreateTxn(accountAddr, note, txParams,
		totalIssuance, decimals, defaultFrozen, manager, reserve, freeze, clawback,
		unitName, assetName, assetViewURL, assetMetadataHash)
	if err != nil {
		fmt.Printf("Failed to make asset: %s\n", err)
		return "",err
	}
	fmt.Printf("Asset created AssetName: %s\n", txn.AssetConfigTxnFields.AssetParams.AssetName)
	// Sign the transaction
	signResponse, err := kmdClient.SignTransaction(exampleWalletHandleToken, walletPassword, txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction with kmd: %s\n", err)
		return "",err
	}
	fmt.Printf("kmd made signed transaction with bytes: %x\n", signResponse.SignedTransaction)

	// Broadcast the transaction to the network
	// Note that this transaction will get rejected because the accounts do not have any tokens
	sendResponse, err := algodClient.SendRawTransaction(signResponse.SignedTransaction)
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return "",err
	}

	fmt.Printf("Transaction ID: %s\n", sendResponse.TxID)
	return sendResponse.TxID, nil
	//found wallet 'mylinuxwallet' with ID: 0377a9874ab7818c861ba465647e7e7b
	//Asset created AssetName: Portrait of a Lady (1/15)
	//kmd made signed transaction with bytes: 82a3736967c440f2745f123d5744c13b6210ab1a74b2f395fc022a9bd738d7cc5f0c34661740c162c4555753bad6a422078bb8b16bb566e4e0da9593a696398927931fdae28d08a374786e89a46170617289a2616dc4203062633737373332396132393139666436666661393662616365396134373739a2616eb9506f727472616974206f662061204c6164792028312f313529a26175be68747470733a2f2f756e636f706965642e6f72672f78456b4d5a2e706e67a163c4201698254309153916d850adad2b10e67807bf848c84d0ca8fd72efc0a21ea75eca166c4201698254309153916d850adad2b10e67807bf848c84d0ca8fd72efc0a21ea75eca16dc4201698254309153916d850adad2b10e67807bf848c84d0ca8fd72efc0a21ea75eca172c4201698254309153916d850adad2b10e67807bf848c84d0ca8fd72efc0a21ea75eca17401a2756ea66c6164792d31a3666565cd03e8a26676ce009db404a367656eac6d61696e6e65742d76312e30a26768c420c061c4d8fc1dbdded2d7604be4568e3f6d041987ac37bde4b620b5ab39248adfa26c76ce009db7eca46e6f7465c4117465737420617373657420637265617465a3736e64c4201698254309153916d850adad2b10e67807bf848c84d0ca8fd72efc0a21ea75eca474797065a461636667
	//failed to send transaction: HTTP 400 Bad Request: TransactionPool.Remember: transaction AJLT4KTFCSFILGVORFMQQPDTVV5PAKQIHCTQZ6SONUY33RBO76YA: account C2MCKQYJCU4RNWCQVWWSWEHGPAD37BEMQTIMVD6XF36AUIPKOXWIZOO7ZE balance 197000 below min 200000 (1 assets)

	//Made a kmd client
	//Made an algod client
	//Got 2 wallet(s):
	//ID: 0377a9874ab7818c861ba465647e7e7b    Name: mylinuxwallet
	//found wallet 'mylinuxwallet' with ID: 0377a9874ab7818c861ba465647e7e7b
	//Asset created AssetName: Portrait of a Lady (1/15)
	//kmd made signed transaction with bytes: 82a3736967c4406808b780e411798800420fe58d3f50f30946ba5f25fcbef955d2fb17de47b2f2ee8ec4f75b3d1b9df3497e7ab8ee27e6b465f9f2a0bc6300baa50aea8f592a0aa374786e89a46170617289a2616dc4203062633737373332396132393139666436666661393662616365396134373739a2616eb9506f727472616974206f662061204c6164792028312f313529a26175be68747470733a2f2f756e636f706965642e6f72672f78456b4d5a2e706e67a163c4201698254309153916d850adad2b10e67807bf848c84d0ca8fd72efc0a21ea75eca166c4201698254309153916d850adad2b10e67807bf848c84d0ca8fd72efc0a21ea75eca16dc4201698254309153916d850adad2b10e67807bf848c84d0ca8fd72efc0a21ea75eca172c4201698254309153916d850adad2b10e67807bf848c84d0ca8fd72efc0a21ea75eca17401a2756ea66c6164792d31a3666565cd03e8a26676ce009db441a367656eac6d61696e6e65742d76312e30a26768c420c061c4d8fc1dbdded2d7604be4568e3f6d041987ac37bde4b620b5ab39248adfa26c76ce009db829a46e6f7465c4117465737420617373657420637265617465a3736e64c4201698254309153916d850adad2b10e67807bf848c84d0ca8fd72efc0a21ea75eca474797065a461636667
	//Transaction ID: TGQ4B7ISZYUKOTU4K2OVYVZLH256UJLCMUWNBVSX6AYEFLXUHGGA
}
