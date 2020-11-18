package main

import (
	"fmt"

	"github.com/algorand/go-algorand-sdk/client/algod"
	"github.com/algorand/go-algorand-sdk/client/kmd"
)

const algodAddress = "http://localhost:8080"
const kmdAddress = "http://localhost:7833"
const algodToken = "contents-of-algod.token"
const kmdToken = "contents-of-kmd.token"

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
