package main

import (
	"fmt"

	"github.com/algorand/go-algorand-sdk/client/algod"
	"github.com/algorand/go-algorand-sdk/client/kmd"
)

// CHANGE ME
const kmdAddress = "http://localhost:7833"
const kmdToken = "206ba3f9ad1d83523fb2a303dd055cd99ce10c5be01e35ee88285fe51438f02a"
const algodAddress = "http://localhost:8080"
const algodToken = "da61ace80780af7b1c78456c7d1d2a511758a754d2c219e1a6b37c32763f5bfe"

func main() {
	// Create an algod client
	algodClient, err := algod.MakeClient(algodAddress, algodToken)
	if err != nil {
		return
	}

	// Create a kmd client
	kmdClient, err := kmd.MakeClient(kmdAddress, kmdToken)
	if err != nil {
		return
	}

	fmt.Printf("algod: %T, kmd: %T\n", algodClient, kmdClient)
}
