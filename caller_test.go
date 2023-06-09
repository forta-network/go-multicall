package multicall

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/forta-network/go-multicall/contracts/contract_multicall"
	"github.com/stretchr/testify/require"
)

type testType struct {
	Val1 bool
	Val2 string
	Val3 []string
	Val4 []*big.Int
	Val5 *big.Int
	Val6 common.Address
}

const (
	testAddr1 = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
	testAddr2 = "0x64d5192F03bD98dB1De2AA8B4abAC5419eaC32CE"
)

const testABI = `[
	{
		"constant":true,
		"inputs":[
			{
				"name":"val1",
				"type":"bool"
			},
			{
				"name":"val2",
				"type":"string"
			},
			{
				"name":"val1",
				"type":"string[]"
			},
			{
				"name":"val4",
				"type":"uint256[]"
			},
			{
				"name":"val5",
				"type":"uint256"
			},
			{
				"name":"val6",
				"type":"address"
			}
		],
		"name":"testFunc",
		"outputs":[
			{
				"name":"val1",
				"type":"bool"
			},
			{
				"name":"val2",
				"type":"string"
			},
			{
				"name":"val1",
				"type":"string[]"
			},
			{
				"name":"val4",
				"type":"uint256[]"
			},
			{
				"name":"val5",
				"type":"uint256"
			},
			{
				"name":"val6",
				"type":"address"
			}
		],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	}
]`

type multicallStub struct {
	returnData func(calls []contract_multicall.Multicall3Call3) [][]byte
}

func (ms *multicallStub) Aggregate3(opts *bind.CallOpts, calls []contract_multicall.Multicall3Call3) (results []contract_multicall.Multicall3Result, err error) {
	allReturnData := ms.returnData(calls)
	for _, returnData := range allReturnData {
		results = append(results, contract_multicall.Multicall3Result{
			Success:    true,
			ReturnData: returnData,
		})
	}
	return
}

func TestCaller_TwoCalls(t *testing.T) {
	r := require.New(t)

	testContract1, err := NewContract(testABI, testAddr1)
	r.NoError(err)

	testContract2, err := NewContract(testABI, testAddr2)
	r.NoError(err)

	values1 := testType{
		Val1: true,
		Val2: "val2",
		Val3: []string{"val3_1", "val3_2"},
		Val4: []*big.Int{big.NewInt(123), big.NewInt(456)},
		Val5: big.NewInt(678),
		Val6: common.HexToAddress(testAddr1),
	}

	call1 := testContract1.NewCall(
		new(testType), "testFunc",
		values1.Val1, values1.Val2, values1.Val3,
		values1.Val4, values1.Val5, values1.Val6,
	)

	values2 := testType{
		Val1: false,
		Val2: "val2_alt",
		Val3: []string{"val3_1_alt", "val3_2_alt"},
		Val4: []*big.Int{big.NewInt(1239), big.NewInt(4569)},
		Val5: big.NewInt(6789),
		Val6: common.HexToAddress(testAddr2),
	}

	call2 := testContract2.NewCall(
		new(testType), "testFunc",
		values2.Val1, values2.Val2, values2.Val3,
		values2.Val4, values2.Val5, values2.Val6,
	)

	caller := &Caller{
		contract: &multicallStub{
			returnData: func(calls []contract_multicall.Multicall3Call3) [][]byte {
				return [][]byte{
					// return inputs as outputs by stripping the method prefix
					calls[0].CallData[4:],
				}
			},
		},
	}

	calls, err := caller.CallChunked(nil, 1, call1, call2)
	r.NoError(err)

	call1Out := calls[0].Outputs.(*testType)
	r.Equal(values1.Val1, call1Out.Val1)
	r.Equal(values1.Val2, call1Out.Val2)
	r.Equal(values1.Val3, call1Out.Val3)
	r.Equal(values1.Val4, call1Out.Val4)
	r.Equal(values1.Val5, call1Out.Val5)
	r.Equal(values1.Val6, call1Out.Val6)

	call2Out := calls[1].Outputs.(*testType)
	r.Equal(values2.Val1, call2Out.Val1)
	r.Equal(values2.Val2, call2Out.Val2)
	r.Equal(values2.Val3, call2Out.Val3)
	r.Equal(values2.Val4, call2Out.Val4)
	r.Equal(values2.Val5, call2Out.Val5)
	r.Equal(values2.Val6, call2Out.Val6)
}

const emptyABI = `[
	{
		"constant":true,
		"inputs": [],
		"name":"testFunc",
		"outputs": [],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	}
]`

func TestCaller_EmptyCall(t *testing.T) {
	r := require.New(t)

	testContract, err := NewContract(emptyABI, testAddr1)
	r.NoError(err)

	call := testContract.NewCall(
		new(struct{}), "testFunc",
		// no inputs
	)

	caller := &Caller{
		contract: &multicallStub{
			returnData: func(calls []contract_multicall.Multicall3Call3) [][]byte {
				return [][]byte{
					// return empty output
					make([]byte, 0),
				}
			},
		},
	}

	calls, err := caller.CallChunked(nil, 1, call)
	r.NoError(err)
	r.Len(calls, 1)
}

const oneValueABI = `[
	{
		"constant":true,
		"inputs": [
			{
				"name":"val1",
				"type":"bool"
			}
		],
		"name":"testFunc",
		"outputs": [
			{
				"name":"val1",
				"type":"bool"
			}
		],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	}
]`

func TestCaller_BadInput(t *testing.T) {
	r := require.New(t)

	testContract, err := NewContract(oneValueABI, testAddr1)
	r.NoError(err)

	call := testContract.NewCall(
		new(struct{}), "testFunc",
		'a',
	)

	caller := &Caller{
		contract: &multicallStub{
			returnData: func(calls []contract_multicall.Multicall3Call3) [][]byte {
				return [][]byte{
					// return bad output
					{},
				}
			},
		},
	}

	calls, err := caller.Call(nil, call)
	r.Error(err)
	r.ErrorContains(err, "cannot use")
	r.Len(calls, 1)
}

func TestCaller_BadOutput(t *testing.T) {
	r := require.New(t)

	testContract, err := NewContract(emptyABI, testAddr1)
	r.NoError(err)

	call := testContract.NewCall(
		new(struct{}), "testFunc",
		// no inputs
	)

	caller := &Caller{
		contract: &multicallStub{
			returnData: func(calls []contract_multicall.Multicall3Call3) [][]byte {
				return [][]byte{
					// return bad output
					{'a'},
				}
			},
		},
	}

	calls, err := caller.Call(nil, call)
	r.Error(err)
	r.Len(calls, 1)
}

func TestCaller_WrongOutputsType(t *testing.T) {
	r := require.New(t)

	testContract, err := NewContract(oneValueABI, testAddr1)
	r.NoError(err)

	call := testContract.NewCall(
		new([]struct{}), "testFunc",
		true,
	)

	packedOutput, err := testContract.ABI.Pack("testFunc", true)
	r.NoError(err)

	caller := &Caller{
		contract: &multicallStub{
			returnData: func(calls []contract_multicall.Multicall3Call3) [][]byte {
				return [][]byte{
					packedOutput,
				}
			},
		},
	}

	calls, err := caller.Call(nil, call)
	r.Error(err)
	r.ErrorContains(err, "not a struct")
	r.Len(calls, 1)
}

func TestDial(t *testing.T) {
	r := require.New(t)

	caller, err := Dial(context.Background(), "https://polygon-rpc.com")
	r.NoError(err)
	r.NotNil(caller)
}
