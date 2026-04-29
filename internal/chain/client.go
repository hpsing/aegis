package chain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// EthClient is the minimal surface verifier code depends on. PublicClient
// (real ethclient) and MockClient (in-memory) both satisfy it.
type EthClient interface {
	// TransactionReceipt fetches the receipt for txHash. Returns
	// ErrTxNotFound (wrapped) if the tx isn't on chain yet.
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	// ChainID returns the chain id of the connected node.
	ChainID(ctx context.Context) (*big.Int, error)
	// BlockNumber returns the latest block number. Verifier uses this to
	// require N confirmations past the receipt's block before voting.
	BlockNumber(ctx context.Context) (uint64, error)
}

// ErrTxNotFound is what callers should check for to distinguish "the tx
// isn't on-chain" from arbitrary RPC errors. PublicClient unwraps the
// underlying ethereum.NotFound into this.
type ErrTxNotFoundType struct{}

func (ErrTxNotFoundType) Error() string { return "tx not found" }

// ErrTxNotFound is a sentinel returned when the receipt is not yet on chain.
var ErrTxNotFound = ErrTxNotFoundType{}
