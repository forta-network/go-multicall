# go-multicall
![coverage](https://img.shields.io/badge/coverage-87.7%25-brightgreen)
![build](https://github.com/forta-network/go-multicall/actions/workflows/build.yml/badge.svg)

A thin Go client for making multiple function calls in single `eth_call` request

- Uses the go-ethereum tools and libraries
- Interfaces with the [MakerDAO `Multicall3` contract](https://github.com/mds1/multicall)

_**Warning:** MakerDAO Multicall contracts are different than the [OpenZeppelin Multicall contract](https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/utils/Multicall.sol). Please see [this thread](https://forum.openzeppelin.com/t/multicall-by-oz-and-makerdao-has-a-difference/9350) in the OpenZeppelin forum if you are looking for an explanation._

## Install

```
go get github.com/forta-network/go-multicall
```

## Example

(See other examples under the `examples` directory!)

#### Multicall   
```go
package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/forta-network/go-multicall"
	"github.com/ethereum/go-ethereum/common"
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
```

#### SingleCall   
```go
package main

import (
	"context"
	"fmt"
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

func main() {
	caller, err := multicall.Dial(context.Background(), APIURL)
	if err != nil {
		panic(err)
	}

	contract, err := multicall.NewContract(ERC20ABI, "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	if err != nil {
		panic(err)
	}

	single, err := caller.CallSingle(nil,
		contract.NewCall(
			nil,
			"balanceOf",
			common.HexToAddress("0xcEe284F754E854890e311e3280b767F80797180d"), // Arbitrum One gateway
		).Name("Arbitrum One gateway balance").SetExtend(map[string]string{
			"account": "0xcEe284F754E854890e311e3280b767F80797180d",
			"token":   "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
		}))
	if err != nil {
		return
	}

	fmt.Println(single.CallName, ":", single.UnpackResult()[0].(*big.Int), single.Extend)
}
```
