package multicall

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCall_BadABI(t *testing.T) {
	r := require.New(t)

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
	` // missing closing ] at the end

	_, err := NewContract(oneValueABI, "0x")
	r.Error(err)
	r.ErrorContains(err, "unexpected EOF")
}
