package multicall

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/forta-network/go-multicall/contracts/contract_multicall"
)

// DefaultAddress is the same for all chains (Multicall3).
// Taken from https://github.com/mds1/multicall
const DefaultAddress = "0xcA11bde05977b3631167028862bE2a173976CA11"

// Caller makes multicalls.
type Caller struct {
	contract contract_multicall.Interface
}

// New creates a new caller.
func New(client bind.ContractCaller, multicallAddr ...string) (*Caller, error) {
	addr := DefaultAddress
	if multicallAddr != nil {
		addr = multicallAddr[0]
	}
	contract, err := contract_multicall.NewMulticallCaller(common.HexToAddress(addr), client)
	if err != nil {
		return nil, err
	}
	return &Caller{
		contract: contract,
	}, nil
}

// Dial dials and Ethereum JSON-RPC API and uses the client as the
// caller backend.
func Dial(ctx context.Context, rawUrl string, multicallAddr ...string) (*Caller, error) {
	client, err := ethclient.DialContext(ctx, rawUrl)
	if err != nil {
		return nil, err
	}
	return New(client, multicallAddr...)
}

// Call makes multicalls.
func (caller *Caller) Call(opts *bind.CallOpts, calls ...*Call) ([]*Call, error) {
	return caller.calls(opts, calls...)
}

func (caller *Caller) CallSingle(opts *bind.CallOpts, call *Call) (*Call, error) {

	calls, err := caller.calls(opts, call)
	if err != nil {
		return call, fmt.Errorf("CallSingle failed: %v", err)
	}
	return calls[0], nil
}

func (caller *Caller) calls(opts *bind.CallOpts, calls ...*Call) ([]*Call, error) {
	var multiCalls []contract_multicall.Multicall3Call3

	for i, call := range calls {
		b, err := call.Pack()
		if err != nil {
			return calls, fmt.Errorf("failed to pack call inputs at index [%d]: %v", i, err)
		}
		multiCalls = append(multiCalls, contract_multicall.Multicall3Call3{
			Target:       call.Contract.Address,
			AllowFailure: call.CanFail,
			CallData:     b,
		})
	}

	results, err := caller.contract.Aggregate3(opts, multiCalls)
	if err != nil {
		return calls, fmt.Errorf("multicall failed: %v", err)
	}

	for i, result := range results {
		call := calls[i] // index always matches
		call.Failed = !result.Success
		if err := call.Unpack(result.ReturnData); err != nil {
			if call.CanFail {
				log.Println(fmt.Errorf("failed to unpack call outputs at index [%d]: %v", i, err))
				continue
			}
			return calls, fmt.Errorf("failed to unpack call outputs at index [%d]: %v", i, err)
		}
	}

	return calls, nil
}

// CallChunked makes multiple multicalls by chunking given calls.
// Cooldown is helpful for sleeping between chunks and avoiding rate limits.
func (caller *Caller) CallChunked(opts *bind.CallOpts, chunkSize int, cooldown time.Duration, calls ...*Call) ([]*Call, error) {
	var allCalls []*Call
	for i, chunk := range chunkInputs(chunkSize, calls) {
		if i > 0 && cooldown > 0 {
			time.Sleep(cooldown)
		}

		chunk, err := caller.calls(opts, chunk...)
		if err != nil {
			return calls, fmt.Errorf("call chunk [%d] failed: %v", i, err)
		}
		allCalls = append(allCalls, chunk...)
	}
	return allCalls, nil
}

func chunkInputs[T any](chunkSize int, inputs []T) (chunks [][]T) {
	if len(inputs) == 0 {
		return
	}

	if chunkSize <= 0 || len(inputs) < 2 || chunkSize > len(inputs) {
		return [][]T{inputs}
	}

	lastChunkSize := len(inputs) % chunkSize

	chunkCount := len(inputs) / chunkSize

	for i := 0; i < chunkCount; i++ {
		start := i * chunkSize
		end := start + chunkSize
		chunks = append(chunks, inputs[start:end])
	}

	if lastChunkSize > 0 {
		start := chunkCount * chunkSize
		end := start + lastChunkSize
		chunks = append(chunks, inputs[start:end])
	}

	return
}
