package main

import (
	"encoding/json"
	"fmt"

	"github.com/algorand/go-algorand-sdk/client/algod"
)

// These constants represent the algod REST endpoint and the corresponding
// API token. You can retrieve these from the `algod.net` and `algod.token`
// files in the algod data directory.
//
//Last committed block: 27292
//Time since last block: 0.0s
//Sync Time: 731.9s
//Last consensus protocol: https://github.com/algorandfoundation/specs/tree/5615adc36bad610c7f165fa2967f4ecfa75125f0
//Next consensus protocol: https://github.com/algorandfoundation/specs/tree/5615adc36bad610c7f165fa2967f4ecfa75125f0
//Round for next consensus protocol: 27293
//Next consensus protocol supported: true
//Last Catchpoint:
//Genesis ID: mainnet-v1.0
//Genesis hash: wGHE2Pwdvd7S12BL5FaOP20EGYesN73ktiC1qzkkit8=

const algodAddress = "http://localhost:8080"
const algodToken = "3005189549ea62dc24fac6102493c941f68840304d3819727b83d3eb74de8e6d"

func main() {
	// Create an algod client
	algodClient, err := algod.MakeClient(algodAddress, algodToken)
	if err != nil {
		fmt.Printf("failed to make algod client: %s\n", err)
		return
	}

	// Print algod status
	nodeStatus, err := algodClient.Status()
	if err != nil {
		fmt.Printf("error getting algod status: %s\n", err)
		return
	}

	fmt.Printf("algod last round: %d\n", nodeStatus.LastRound)
	fmt.Printf("algod time since last round: %d\n", nodeStatus.TimeSinceLastRound)
	fmt.Printf("algod catchup: %d\n", nodeStatus.CatchupTime)
	fmt.Printf("algod latest version: %s\n", nodeStatus.LastVersion)

	// Fetch block information
	lastBlock, err := algodClient.Block(nodeStatus.LastRound)
	if err != nil {
		fmt.Printf("error getting last block: %s\n", err)
		return
	}

	// Print the block information
	fmt.Printf("\n-----------------Block Information-------------------\n")
	blockJSON, err := json.MarshalIndent(lastBlock, "", "\t")
	if err != nil {
		fmt.Printf("Can not marshall block data: %s\n", err)
	}
	fmt.Printf("%s\n", blockJSON)
}
