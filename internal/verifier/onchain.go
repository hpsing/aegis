package verifier

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hpsing/aegis/internal/aegis"
	"github.com/hpsing/aegis/internal/keeperhub"
)

// OnChain is the high-level wrapper a verifier uses to interact with
// AegisContract + VerifierRegistry. ALL writes (commit, reveal, settle)
// go through `keeper` — the KeeperHub.Client — instead of broadcasting
// directly. This satisfies the architecture-doc §4 invariant ("every
// settlement-critical tx routes through KeeperHub for retry, gas
// optimization, audit trail").
type OnChain struct {
	Aegis     *aegis.AegisContract
	Registry  *aegis.VerifierRegistry
	signerKey *ecdsa.PrivateKey
	chainID   *big.Int
	client    *ethclient.Client
	keeper    keeperhub.Client
	aegisAddr common.Address
}

// NewOnChain wires bindings to the given RPC + key. `keeper` is the
// KeeperHub.Client every write routes through; tests pass a MockClient,
// production passes a LiveClient.
func NewOnChain(
	ctx context.Context,
	rpcURL string,
	aegisAddr, registryAddr common.Address,
	signerKey *ecdsa.PrivateKey,
	keeper keeperhub.Client,
) (*OnChain, error) {
	if keeper == nil {
		return nil, fmt.Errorf("keeperhub.Client required")
	}
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
	return &OnChain{
		Aegis:     q,
		Registry:  r,
		signerKey: signerKey,
		chainID:   chainID,
		client:    client,
		keeper:    keeper,
		aegisAddr: aegisAddr,
	}, nil
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

// noSendTransactor builds calldata via abigen without broadcasting. We
// then forward the (to, data, gas) tuple to KeeperHub.Client which
// signs+sends with retry/gas-bump/audit-trail.
func (o *OnChain) noSendTransactor(ctx context.Context) (*bind.TransactOpts, error) {
	opts, err := o.transactor(ctx)
	if err != nil {
		return nil, err
	}
	opts.NoSend = true
	return opts, nil
}

// CommitVote computes the commit hash and routes the commitVote()
// calldata through KeeperHub.
func (o *OnChain) CommitVote(
	ctx context.Context, jobID *big.Int, verdict bool, nonce [32]byte,
) (keeperhub.Receipt, error) {
	opts, err := o.noSendTransactor(ctx)
	if err != nil {
		return keeperhub.Receipt{}, err
	}
	hash := commitHash(verdict, nonce, o.signerAddress())
	tx, err := o.Aegis.CommitVote(opts, jobID, hash)
	if err != nil {
		return keeperhub.Receipt{}, fmt.Errorf("build commit calldata: %w", err)
	}
	return o.send(ctx, tx, jobID, keeperhub.PurposeCommit)
}

// RevealVote routes revealVote() through KeeperHub.
func (o *OnChain) RevealVote(
	ctx context.Context, jobID *big.Int, verdict bool, nonce [32]byte,
) (keeperhub.Receipt, error) {
	opts, err := o.noSendTransactor(ctx)
	if err != nil {
		return keeperhub.Receipt{}, err
	}
	tx, err := o.Aegis.RevealVote(opts, jobID, verdict, nonce)
	if err != nil {
		return keeperhub.Receipt{}, fmt.Errorf("build reveal calldata: %w", err)
	}
	return o.send(ctx, tx, jobID, keeperhub.PurposeReveal)
}

// Settle routes settle() through KeeperHub.
func (o *OnChain) Settle(ctx context.Context, jobID *big.Int) (keeperhub.Receipt, error) {
	opts, err := o.noSendTransactor(ctx)
	if err != nil {
		return keeperhub.Receipt{}, err
	}
	tx, err := o.Aegis.Settle(opts, jobID)
	if err != nil {
		return keeperhub.Receipt{}, fmt.Errorf("build settle calldata: %w", err)
	}
	return o.send(ctx, tx, jobID, keeperhub.PurposeSettle)
}

// send is the shared "abigen tx → KeeperHub" handoff. NoSend transactors
// always populate `*types.Transaction.To/Data/Gas` even though they don't
// broadcast, so we extract those fields and pass them on.
func (o *OnChain) send(
	ctx context.Context,
	tx interface {
		To() *common.Address
		Data() []byte
		Value() *big.Int
		Gas() uint64
	},
	jobID *big.Int,
	purpose keeperhub.Purpose,
) (keeperhub.Receipt, error) {
	to := tx.To()
	if to == nil {
		return keeperhub.Receipt{}, fmt.Errorf("nil recipient on built tx")
	}
	return o.keeper.SendTransaction(ctx, keeperhub.SendTxInput{
		ChainID:  o.chainID.Uint64(),
		To:       *to,
		Data:     tx.Data(),
		Value:    tx.Value(),
		GasLimit: tx.Gas(),
		Metadata: keeperhub.Metadata{
			JobID:          jobID.Uint64(),
			Purpose:        purpose,
			IdempotencyKey: keeperhub.DefaultIdempotencyKey(jobID.Uint64(), purpose),
		},
	})
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
