package uniswapv3

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// SwapEventJSON is the ABI fragment for IUniswapV3Pool.Swap. Pasted as a
// JSON literal so we don't have to ship the whole pool ABI binary.
//
// event Swap(
//
//	address indexed sender,
//	address indexed recipient,
//	int256  amount0,
//	int256  amount1,
//	uint160 sqrtPriceX96,
//	uint128 liquidity,
//	int24   tick
//
// )
const SwapEventJSON = `[
  {
    "anonymous": false,
    "inputs": [
      { "indexed": true,  "internalType": "address", "name": "sender",      "type": "address" },
      { "indexed": true,  "internalType": "address", "name": "recipient",   "type": "address" },
      { "indexed": false, "internalType": "int256",  "name": "amount0",     "type": "int256"  },
      { "indexed": false, "internalType": "int256",  "name": "amount1",     "type": "int256"  },
      { "indexed": false, "internalType": "uint160", "name": "sqrtPriceX96","type": "uint160" },
      { "indexed": false, "internalType": "uint128", "name": "liquidity",   "type": "uint128" },
      { "indexed": false, "internalType": "int24",   "name": "tick",        "type": "int24"   }
    ],
    "name": "Swap",
    "type": "event"
  }
]`

// SwapEventABI is the parsed ABI containing just the Swap event.
var SwapEventABI abi.ABI

// SwapEventID is topic[0] of any Uniswap V3 Swap log.
var SwapEventID common.Hash

func init() {
	parsed, err := abi.JSON(strings.NewReader(SwapEventJSON))
	if err != nil {
		// Static literal — must never fail. Panic in init is appropriate
		// since the package is unusable otherwise.
		panic("uniswapv3: parse Swap ABI: " + err.Error())
	}
	SwapEventABI = parsed
	SwapEventID = parsed.Events["Swap"].ID
}
