package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/forta-network/go-multicall"
)

const (
	APIURL   = "https://cloudflare-eth.com"
	ERC20ABI = `[
		{
			"constant":true,
			"inputs":[
					{
						"name":"tokenOwner",
						"type":"address"
					}
			],
			"name":"balanceOf",
			"outputs":[
					{
						"name":"balance",
						"type":"uint256"
					}
			],
			"payable":false,
			"stateMutability":"view",
			"type":"function"
		}
	]`
)

type balanceOutput struct {
	Balance *big.Int
}

func main() {
	caller, err := multicall.Dial(context.Background(), APIURL)
	if err != nil {
		panic(err)
	}

	contract, err := multicall.NewContract(ERC20ABI, "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	if err != nil {
		panic(err)
	}

	calls, err := caller.Call(nil,
		contract.NewCall(
			new(balanceOutput),
			"balanceOf",
			common.HexToAddress("0xcEe284F754E854890e311e3280b767F80797180d"), // Arbitrum One gateway
		).Name("Arbitrum One gateway balance"),
		contract.NewCall(
			new(balanceOutput),
			"balanceOf",
			common.HexToAddress("0x40ec5B33f54e0E8A33A975908C5BA1c14e5BbbDf"), // Polygon ERC20 bridge
		).Name("Polygon ERC20 bridge balance"),
	)
	if err != nil {
		panic(err)
	}
	for _, call := range calls {
		fmt.Println(call.CallName, ":", call.Outputs.(*balanceOutput).Balance)
	}
}
