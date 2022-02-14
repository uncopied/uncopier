package main

import (
	"encoding/json"
	"fmt"
	"github.com/algorand/go-algorand-sdk/client/algod"
	"github.com/algorand/go-algorand-sdk/client/kmd"
	transaction "github.com/algorand/go-algorand-sdk/future"
	"io"
	"log"
	"net/http"
	"strings"
)

type UncopiedAccount struct {
	Account struct {
		Address                     string `json:"address"`
		Amount                      int    `json:"amount"`
		AmountWithoutPendingRewards int    `json:"amount-without-pending-rewards"`
		Assets                      []struct {
			Amount         int    `json:"amount"`
			AssetID        int    `json:"asset-id"`
			Creator        string `json:"creator"`
			Deleted        bool   `json:"deleted"`
			IsFrozen       bool   `json:"is-frozen"`
			OptedInAtRound int    `json:"opted-in-at-round"`
		} `json:"assets"`
		CreatedAssets []struct {
			CreatedAtRound int  `json:"created-at-round"`
			Deleted        bool `json:"deleted"`
			Index          int  `json:"index"`
			Params         struct {
				Clawback      string `json:"clawback"`
				Creator       string `json:"creator"`
				Decimals      int    `json:"decimals"`
				DefaultFrozen bool   `json:"default-frozen"`
				Freeze        string `json:"freeze"`
				Manager       string `json:"manager"`
				MetadataHash  string `json:"metadata-hash"`
				Name          string `json:"name"`
				NameB64       string `json:"name-b64"`
				Reserve       string `json:"reserve"`
				Total         int    `json:"total"`
				UnitName      string `json:"unit-name"`
				UnitNameB64   string `json:"unit-name-b64"`
				URL           string `json:"url"`
				URLB64        string `json:"url-b64"`
			} `json:"params"`
		} `json:"created-assets"`
		CreatedAtRound int  `json:"created-at-round"`
		Deleted        bool `json:"deleted"`
		Participation  struct {
			SelectionParticipationKey string `json:"selection-participation-key"`
			VoteFirstValid            int    `json:"vote-first-valid"`
			VoteKeyDilution           int    `json:"vote-key-dilution"`
			VoteLastValid             int    `json:"vote-last-valid"`
			VoteParticipationKey      string `json:"vote-participation-key"`
		} `json:"participation"`
		PendingRewards int    `json:"pending-rewards"`
		RewardBase     int    `json:"reward-base"`
		Rewards        int    `json:"rewards"`
		Round          int    `json:"round"`
		SigType        string `json:"sig-type"`
		Status         string `json:"status"`
	} `json:"account"`
	CurrentRound int `json:"current-round"`
}

func main() {

	kmdAddress := "http://91.121.222.154:7833"
	kmdToken := "f457576dea5d6fb59894e2d9fabeffec316fd81406a50e72704dff39da019eaa"
	algodAddress := "http://91.121.222.154:8888"
	algodToken := "3090f291739c0d3af6c5f53fa56bbd9f1fd90f400315c1473e572e85e37edf2e"
	walletName := "uncopied_art"
	walletPassword := "Ekj3M#KmUnco#Art"
	accountAddr := "42K2TG6IVACAZSAPOQUZANFINN5DVJGIJLTVTREXAAACN4KUNMBJFD7JFI"

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
		return
	}

	// Extract the wallet handle
	exampleWalletHandleToken := initResponse.WalletHandleToken

	url := "https://algoindexer.algoexplorerapi.io/v2/accounts/42K2TG6IVACAZSAPOQUZANFINN5DVJGIJLTVTREXAAACN4KUNMBJFD7JFI?include-all=false"
	var client http.Client
	resp, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		//bodyString := string(bodyBytes)
		//fmt.Printf("here is the respons %s", bodyString)
		var uncopiedAccount UncopiedAccount
		err = json.Unmarshal(bodyBytes, &uncopiedAccount)
		if err != nil {
			fmt.Errorf("oops ", err)
		}
		for i, a := range uncopiedAccount.Account.CreatedAssets {
			if strings.Contains(a.Params.URL, "uncopied.org") && !a.Deleted {
				fmt.Printf("here is the object to DESTROY %v %v name %v url : %v \n", i, a.Index, a.Params.Name, a.Params.URL)

				assetID := uint64(a.Index)
				// Get network-related transaction parameters and assign
				// Get the suggested transaction parameters
				txParams, err := algodClient.BuildSuggestedParams()
				if err != nil {
					fmt.Printf("error getting suggested tx params: %s\n", err)
					return
				}

				// comment out the next two (2) lines to use suggested fees
				//txParams.FlatFee = true
				//txParams.Fee = 1000
				note := "2022-02-14 : destroy uncopied.org assets"

				txn, err := transaction.MakeAssetDestroyTxn(accountAddr, []byte(note), txParams, assetID)
				if err != nil {
					fmt.Printf("Failed to send txn: %s\n", err)
					return
				} else {
					fmt.Printf("Prepared txn: %v\n", txn)
				}

				fmt.Printf("Asset destroy AssetID: %v\n", assetID)
				// Sign the transaction
				signResponse, err := kmdClient.SignTransaction(exampleWalletHandleToken, walletPassword, txn)
				if err != nil {
					fmt.Printf("Failed to sign transaction with kmd: %s\n", err)
					return
				} else {
					fmt.Printf("Prepared sign txn: %v\n", signResponse)
				}

				fmt.Printf("kmd made signed transaction with bytes: %x\n", signResponse.SignedTransaction)

				// Broadcast the transaction to the network
				// Note that this transaction will get rejected because the accounts do not have any tokens
				sendResponse, err := algodClient.SendRawTransaction(signResponse.SignedTransaction)
				if err != nil {
					fmt.Printf("failed to send transaction: %s\n", err)
				} else {
					fmt.Printf("Transaction ID: %s\n", sendResponse.TxID)
				}

			} else {
				fmt.Printf("here is the object to KEEP %v %v name %v url : %v \n", i, a.Index, a.Params.Name, a.Params.URL)
			}
		}
	}
}
