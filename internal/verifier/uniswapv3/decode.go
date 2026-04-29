package uniswapv3

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// SwapEvent is the decoded form of Uniswap V3's Swap log.
//
// Sign convention: amount0 and amount1 are SIGNED. Negative means the pool
// sent that token (the swapper received it); positive means the pool
// received it (the swapper sent it). Caller decides which of the two is
// the "amount out" based on token ordering in the pool.
type SwapEvent struct {
	PoolAddress  common.Address
	Sender       common.Address
	Recipient    common.Address
	Amount0      *big.Int // signed
	Amount1      *big.Int // signed
	SqrtPriceX96 *big.Int
	Liquidity    *big.Int
	Tick         *big.Int // signed (int24)
}

var ErrNotSwapLog = errors.New("uniswapv3: log is not a Swap event")

// DecodeSwap parses a single log into a SwapEvent. Returns ErrNotSwapLog
// if topic[0] doesn't match the Swap event id.
func DecodeSwap(log *types.Log) (SwapEvent, error) {
	if len(log.Topics) == 0 || log.Topics[0] != SwapEventID {
		return SwapEvent{}, ErrNotSwapLog
	}
	if len(log.Topics) != 3 {
		return SwapEvent{}, fmt.Errorf("uniswapv3: expected 3 topics, got %d", len(log.Topics))
	}

	indexedArgs := abi.Arguments{
		SwapEventABI.Events["Swap"].Inputs[0], // sender
		SwapEventABI.Events["Swap"].Inputs[1], // recipient
	}
	indexedValues := make(map[string]any)
	if err := abi.ParseTopicsIntoMap(indexedValues, indexedArgs, log.Topics[1:]); err != nil {
		return SwapEvent{}, fmt.Errorf("parse topics: %w", err)
	}

	dataValues, err := SwapEventABI.Unpack("Swap", log.Data)
	if err != nil {
		return SwapEvent{}, fmt.Errorf("unpack data: %w", err)
	}
	if len(dataValues) != 5 {
		return SwapEvent{}, fmt.Errorf("expected 5 data values, got %d", len(dataValues))
	}

	out := SwapEvent{
		PoolAddress:  log.Address,
		Sender:       indexedValues["sender"].(common.Address),
		Recipient:    indexedValues["recipient"].(common.Address),
		Amount0:      dataValues[0].(*big.Int),
		Amount1:      dataValues[1].(*big.Int),
		SqrtPriceX96: dataValues[2].(*big.Int),
		Liquidity:    new(big.Int).SetUint64(uint64(0)),
		Tick:         dataValues[4].(*big.Int),
	}
	// liquidity is uint128 — go-ethereum decodes as *big.Int directly.
	out.Liquidity = dataValues[3].(*big.Int)
	return out, nil
}

// FindSwapInReceipt returns the FIRST Swap log emitted by `pool` in the
// receipt. Returns ErrNotSwapLog wrapped if no matching log exists.
// (Multi-hop swaps emit multiple Swap events in one tx; for hackathon
// scope we only look at the matched pool.)
func FindSwapInReceipt(receipt *types.Receipt, pool common.Address) (SwapEvent, error) {
	for _, log := range receipt.Logs {
		if log.Address != pool {
			continue
		}
		ev, err := DecodeSwap(log)
		if err != nil {
			if errors.Is(err, ErrNotSwapLog) {
				continue
			}
			return SwapEvent{}, err
		}
		return ev, nil
	}
	return SwapEvent{}, fmt.Errorf("%w: no Swap log from pool %s", ErrNotSwapLog, pool.Hex())
}
