package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hpsing/aegis/internal/aegis"
)

// IERC20Approver is the minimal ERC-20 surface PostJob needs. The abigen
// MockUSDC binding satisfies it; in production this is canonical USDC. TODO::
type IERC20Approver interface {
	Approve(opts *bind.TransactOpts, spender common.Address, value *big.Int) (*types.Transaction, error)
}

// PostJob approves USDC and calls AegisContract.postJob. Returns the tx
// of the postJob call (jobId is recovered from the receipt's events
// downstream by the caller, since abigen's bind doesn't return decoded
// receipt-event logs synchronously).
func PostJob(
	ctx context.Context,
	usdc IERC20Approver,
	q *aegis.AegisContract,
	signerKey *ecdsa.PrivateKey,
	chainID *big.Int,
	aegisAddr common.Address,
	executor common.Address,
	specHash common.Hash,
	executorReimbursement, executorFee, verifierBounty *big.Int,
) (*types.Transaction, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(signerKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("transactor: %w", err)
	}
	opts.Context = ctx

	total := new(big.Int).Add(executorReimbursement, executorFee)
	total.Add(total, verifierBounty)

	if usdc != nil {
		approveTx, err := usdc.Approve(opts, aegisAddr, total)
		if err != nil {
			return nil, fmt.Errorf("approve: %w", err)
		}
		_ = approveTx
	}

	// Re-fetch a fresh transactor so the nonce advances past the approve.
	opts2, err := bind.NewKeyedTransactorWithChainID(signerKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("transactor 2: %w", err)
	}
	opts2.Context = ctx

	postTx, err := q.PostJob(opts2, specHash, executor, executorReimbursement, executorFee, verifierBounty)
	if err != nil {
		return nil, fmt.Errorf("postJob: %w", err)
	}
	return postTx, nil
}
