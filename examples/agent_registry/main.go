package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/forta-network/go-multicall"
)

const (
	APIURL           = "https://polygon-rpc.com"
	AgentRegistryABI = `[
		{
			"inputs":[
					{
						"internalType":"uint256",
						"name":"agentId",
						"type":"uint256"
					}
			],
			"name":"getAgentState",
			"outputs":[
					{
						"internalType":"bool",
						"name":"registered",
						"type":"bool"
					},
					{
						"internalType":"address",
						"name":"owner",
						"type":"address"
					},
					{
						"internalType":"uint256",
						"name":"agentVersion",
						"type":"uint256"
					},
					{
						"internalType":"string",
						"name":"metadata",
						"type":"string"
					},
					{
						"internalType":"uint256[]",
						"name":"chainIds",
						"type":"uint256[]"
					},
					{
						"internalType":"bool",
						"name":"enabled",
						"type":"bool"
					},
					{
						"internalType":"uint256",
						"name":"disabledFlags",
						"type":"uint256"
					}
			],
			"stateMutability":"view",
			"type":"function"
		}
	]`
)

type agentState struct {
	Registered    bool
	Owner         common.Address
	AgentVersion  *big.Int
	Metadata      string
	ChainIds      []*big.Int
	Enabled       bool
	DisabledFlags *big.Int
}

func main() {
	caller, err := multicall.Dial(context.Background(), APIURL)
	if err != nil {
		panic(err)
	}

	// Forta AgentRegistry
	agentReg, err := multicall.NewContract(AgentRegistryABI, "0x61447385B019187daa48e91c55c02AF1F1f3F863")
	if err != nil {
		panic(err)
	}

	calls, err := caller.Call(nil,
		agentReg.NewCall(
			new(agentState),
			"getAgentState",
			botHexToBigInt("0x80ed808b586aeebe9cdd4088ea4dea0a8e322909c0e4493c993e060e89c09ed1"),
		),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("owner:", calls[0].Outputs.(*agentState).Owner.String())

	b, _ := json.MarshalIndent(calls[0].Outputs.(*agentState), "", "	")
	fmt.Println(string(b))
}

func botHexToBigInt(hex string) *big.Int {
	return common.HexToHash(hex).Big()
}
