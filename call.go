package multicall

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Contract wraps the parsed ABI and acts as a call factory.
type Contract struct {
	ABI     *abi.ABI
	Address common.Address
}

// NewContract creates a new call factory.
func NewContract(rawJson, address string) (*Contract, error) {
	parsedABI, err := ParseABI(rawJson)
	if err != nil {
		return nil, err
	}
	return &Contract{
		ABI:     parsedABI,
		Address: common.HexToAddress(address),
	}, nil
}

// ParseABI parses raw ABI JSON.
func ParseABI(rawJson string) (*abi.ABI, error) {
	parsed, err := abi.JSON(bytes.NewBufferString(rawJson))
	if err != nil {
		return nil, fmt.Errorf("failed to parse abi: %v", err)
	}
	return &parsed, nil
}

// Call wraps a multicall call.
type Call struct {
	CallName string
	Contract *Contract
	Method   string
	Extend   any
	Inputs   []any
	Outputs  any
	CanFail  bool
	Failed   bool
}

// NewCall creates a new call using given inputs.
// Outputs type is the expected output struct to unpack and set values in.
func (contract *Contract) NewCall(
	outputs any, methodName string, inputs ...any,
) *Call {
	return &Call{
		Contract: contract,
		Method:   methodName,
		Inputs:   inputs,
		Outputs:  outputs,
	}
}

// Name sets a name for the call.
func (call *Call) Name(name string) *Call {
	call.CallName = name
	return call
}

func (call *Call) SetExtend(ext any) *Call {
	call.Extend = ext
	return call
}

func (call *Call) UnpackResult() []interface{} {
	if call.Outputs == nil {
		return nil
	}
	return call.Outputs.([]interface{})
}

// AllowFailure sets if the call is allowed to fail. This helps avoiding a revert
// when one of the calls in the array fails.
func (call *Call) AllowFailure() *Call {
	call.CanFail = true
	return call
}

// Unpack unpacks and converts EVM outputs and sets struct fields.
func (call *Call) Unpack(b []byte) error {
	out, err := call.Contract.ABI.Unpack(call.Method, b)
	if err != nil {
		return fmt.Errorf("failed to unpack '%s' outputs: %v", call.Method, err)
	}
	if call.Outputs == nil {
		call.Outputs = out
		return nil
	}

	t := reflect.ValueOf(call.Outputs)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return errors.New("outputs type is not a struct")
	}

	fieldCount := t.NumField()
	for i := 0; i < fieldCount; i++ {
		field := t.Field(i)
		converted := abi.ConvertType(out[i], field.Interface())
		field.Set(reflect.ValueOf(converted))
	}

	return nil
}

// Pack converts and packs EVM inputs.
func (call *Call) Pack() ([]byte, error) {
	b, err := call.Contract.ABI.Pack(call.Method, call.Inputs...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack '%s' inputs: %v", call.Method, err)
	}
	return b, nil
}
