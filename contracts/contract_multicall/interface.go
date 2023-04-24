package contract_multicall

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// Interface is an abstraction of the contract.
type Interface interface {
	Aggregate3(opts *bind.CallOpts, calls []Multicall3Call3) ([]Multicall3Result, error)
}
