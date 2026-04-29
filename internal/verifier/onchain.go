package verifier

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"

	"github.com/hpsing/aegis/internal/aegis"
)

// OnChain is the high-level wrapper a verifier uses to interact with
// AegisContract + VerifierRegistry. It bundles the abigen bindings
// with a TransactOpts pre-configured from the verifier's private key.
type OnChain struct {
	Aegis     *aegis.AegisContract
	Registry  *aegis.VerifierRegistry
	signerKey *ecdsa.PrivateKey
	chainID   *big.Int
	client    *ethclient.Client
}

// NewOnChain wires bindings to the given RPC + key.
func NewOnChain(
	ctx context.Context,
	rpcURL string,
	aegisAddr, registryAddr common.Address,
	signerKey *ecdsa.PrivateKey,
) (*OnChain, error) {
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", rpcURL, err)
	}
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("chain id: %w", err)
	}
	q, err := aegis.NewAegisContract(aegisAddr, client)
	if err != nil {
		return nil, fmt.Errorf("bind AegisContract: %w", err)
	}
	r, err := aegis.NewVerifierRegistry(registryAddr, client)
	if err != nil {
		return nil, fmt.Errorf("bind VerifierRegistry: %w", err)
	}
	return &OnChain{Aegis: q, Registry: r, signerKey: signerKey, chainID: chainID, client: client}, nil
}

// Address returns the EOA the OnChain client signs as.
func (o *OnChain) Address() common.Address {
	return signerAddrFrom(&o.signerKey.PublicKey)
}

// signerAddress is the internal name (kept for readability at call sites
// that pre-date Address being public).
func (o *OnChain) signerAddress() common.Address {
	return o.Address()
}

// transactor builds a fresh TransactOpts. We rebuild per-call so the nonce
// is fetched fresh and the context can carry per-call timeouts.
func (o *OnChain) transactor(ctx context.Context) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(o.signerKey, o.chainID)
	if err != nil {
		return nil, fmt.Errorf("transactor: %w", err)
	}
	opts.Context = ctx
	return opts, nil
}

// CommitVote computes the commit hash and submits commitVote().
func (o *OnChain) CommitVote(
	ctx context.Context, jobID *big.Int, verdict bool, nonce [32]byte,
) (*types.Transaction, error) {
	opts, err := o.transactor(ctx)
	if err != nil {
		return nil, err
	}
	hash := commitHash(verdict, nonce, o.signerAddress())
	return o.Aegis.CommitVote(opts, jobID, hash)
}

// RevealVote submits revealVote() with the previously committed (verdict, nonce).
func (o *OnChain) RevealVote(
	ctx context.Context, jobID *big.Int, verdict bool, nonce [32]byte,
) (*types.Transaction, error) {
	opts, err := o.transactor(ctx)
	if err != nil {
		return nil, err
	}
	return o.Aegis.RevealVote(opts, jobID, verdict, nonce)
}

// Settle calls settle() — anyone can call once the reveal deadline passes.
func (o *OnChain) Settle(ctx context.Context, jobID *big.Int) (*types.Transaction, error) {
	opts, err := o.transactor(ctx)
	if err != nil {
		return nil, err
	}
	return o.Aegis.Settle(opts, jobID)
}

// Job is the read-side representation of AegisContract.jobs(jobId). Mirrors
// the abigen-generated struct shape but with friendlier field names.
type Job struct {
	Client                common.Address
	Executor              common.Address
	ExecutorReimbursement *big.Int
	ExecutorFee           *big.Int
	VerifierBounty        *big.Int
	SpecHash              [32]byte
	TxHash                [32]byte
	ReportedOutcomeHash   [32]byte
	ClaimDeadline         uint64
	CommitDeadline        uint64
	RevealDeadline        uint64
	Status                uint8
}

// GetJob reads the on-chain Job by id.
func (o *OnChain) GetJob(ctx context.Context, jobID *big.Int) (Job, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	raw, err := o.Aegis.Jobs(callOpts, jobID)
	if err != nil {
		return Job{}, fmt.Errorf("jobs: %w", err)
	}
	return Job{
		Client:                raw.Client,
		Executor:              raw.Executor,
		ExecutorReimbursement: raw.ExecutorReimbursement,
		ExecutorFee:           raw.ExecutorFee,
		VerifierBounty:        raw.VerifierBounty,
		SpecHash:              raw.SpecHash,
		TxHash:                raw.TxHash,
		ReportedOutcomeHash:   raw.ReportedOutcomeHash,
		ClaimDeadline:         raw.ClaimDeadline,
		CommitDeadline:        raw.CommitDeadline,
		RevealDeadline:        raw.RevealDeadline,
		Status:                raw.Status,
	}, nil
}

// WatchJobPosted subscribes to JobPosted events. The returned Subscription
// MUST be `.Unsubscribe()`d on shutdown to release the websocket.
func (o *OnChain) WatchJobPosted(
	ctx context.Context, sink chan<- *aegis.AegisContractJobPosted,
) (event.Subscription, error) {
	return o.Aegis.WatchJobPosted(&bind.WatchOpts{Context: ctx}, sink, nil, nil, nil)
}

// WatchClaimSubmitted — same shape.
func (o *OnChain) WatchClaimSubmitted(
	ctx context.Context, sink chan<- *aegis.AegisContractClaimSubmitted,
) (event.Subscription, error) {
	return o.Aegis.WatchClaimSubmitted(&bind.WatchOpts{Context: ctx}, sink, nil)
}

// WatchJobSettled — same shape.
func (o *OnChain) WatchJobSettled(
	ctx context.Context, sink chan<- *aegis.AegisContractJobSettled,
) (event.Subscription, error) {
	return o.Aegis.WatchJobSettled(&bind.WatchOpts{Context: ctx}, sink, nil)
}

// HeadTimestamp returns block.timestamp of the latest block (Unix seconds).
// Used by the verifier loop's reveal scheduler — polling this instead of
// wall-clock so fast-forwarded chain time (in tests via evm_increaseTime,
// or in production during chain stalls) is detected promptly.
func (o *OnChain) HeadTimestamp(ctx context.Context) (uint64, error) {
	head, err := o.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("head: %w", err)
	}
	return head.Time, nil
}

func (o *OnChain) Close() {
	if o.client != nil {
		o.client.Close()
	}
}
