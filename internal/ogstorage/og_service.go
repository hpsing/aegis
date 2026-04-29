package ogstorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/0gfoundation/0g-storage-client/common"
	"github.com/0gfoundation/0g-storage-client/common/blockchain"
	"github.com/0gfoundation/0g-storage-client/core"
	"github.com/0gfoundation/0g-storage-client/indexer"
	"github.com/0gfoundation/0g-storage-client/transfer"
	"github.com/openweb3/web3go"
	"github.com/sirupsen/logrus"
)

// LiveStorage is the production Storage backed by the real 0G Storage
// Go SDK (github.com/0gfoundation/0g-storage-client).
//
// Uploads happen via core.DataInMemory (no temp files needed), routed
// through the configured indexer + EVM RPC. Each Upload* call:
//
//  1. Marshals the payload to JSON
//  2. Wraps the bytes as core.IterableData
//  3. Calls indexer.SplitableUpload — submits an on-chain tx to register
//     the data, then transfers it to storage nodes selected by the
//     indexer
//  4. Returns the root merkle hash as a "0g://<root>" URI
//
// GetAccuracyHistory is intentionally unimplemented: 0G Storage doesn't
// natively index uploads by author. Callers that need the auditor
// "replay the Log" view should subscribe to QuorumContract events
// off-chain (or use Registry.getAccuracy() as the canonical accuracy
// source). Step 9 will optionally add an on-chain root registry to
// re-enable replay if the demo needs it.
type LiveStorage struct {
	web3    *web3go.Client
	indexer *indexer.Client
}

// NewLiveStorage dials the 0G Galileo EVM RPC + indexer with the given
// private key. The caller is responsible for funding the wallet (faucet
// at https://faucet.0g.ai) before any Upload* call.
func NewLiveStorage(rpcURL, indexerURL, privateKeyHex string) (*LiveStorage, error) {
	if rpcURL == "" || indexerURL == "" {
		return nil, errors.New("ogstorage: rpcURL and indexerURL required")
	}
	if privateKeyHex == "" {
		return nil, errors.New("ogstorage: privateKeyHex required (set OG_PRIVATE_KEY)")
	}
	w3, err := blockchain.NewWeb3(rpcURL, privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("ogstorage: dial 0G EVM RPC: %w", err)
	}

	// IMPORTANT: pass a logger with Level >= WarnLevel. The SDK's
	// progress Reminder calls Log(reminder.logger.Level, ...) which
	// triggers logrus's hard-coded panic when level <= PanicLevel (0).
	// Default LogOption{} would set LogLevel=PanicLevel (zero value),
	// so retries on "log entry unavailable yet" would crash the process.
	sdkLogger := logrus.New()
	sdkLogger.SetOutput(io.Discard)
	sdkLogger.SetLevel(logrus.WarnLevel)

	idx, err := indexer.NewClient(indexerURL, indexer.IndexerClientOption{
		LogOption: common.LogOption{Logger: sdkLogger},
	})
	if err != nil {
		w3.Close()
		return nil, fmt.Errorf("ogstorage: dial indexer: %w", err)
	}
	return &LiveStorage{web3: w3, indexer: idx}, nil
}

// Close releases the web3 client. Safe to call multiple times.
func (l *LiveStorage) Close() {
	if l.web3 != nil {
		l.web3.Close()
		l.web3 = nil
	}
}

// uploadJSON is the shared upload path. Returns "0g://<root-hex>" on success.
func (l *LiveStorage) uploadJSON(ctx context.Context, payload any) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("ogstorage: marshal: %w", err)
	}
	data, err := core.NewDataInMemory(body)
	if err != nil {
		return "", fmt.Errorf("ogstorage: wrap data: %w", err)
	}

	// Our payloads are tiny (sub-KB JSON); 4 GiB fragment size ensures
	// the upload is a single fragment.
	const fragmentSize = int64(4 * 1024 * 1024 * 1024)
	opt := transfer.UploadOption{
		ExpectedReplica:  1,
		TaskSize:         10,
		FinalityRequired: transfer.TransactionPacked,
		// FastMode=false: wait for the on-chain receipt + storage-node
		// log-sync before pushing data. FastMode=true uploads "by root"
		// concurrently with the chain submission, but storage nodes on
		// Galileo lag chain by a few blocks — the upload then races a
		// log-sync the nodes can't satisfy yet, the SDK retries via a
		// Reminder, and (because of an SDK quirk where Reminder.Remind
		// calls Log at the logger's configured level) it panics. Slower
		// but reliable wins.
		FastMode:    false,
		Method:      "min",
		FullTrusted: true,
		NRetries:    5,
	}
	_, roots, err := l.indexer.SplitableUpload(ctx, l.web3, data, fragmentSize, opt)
	if err != nil {
		return "", fmt.Errorf("ogstorage: upload: %w", err)
	}
	if len(roots) == 0 {
		return "", errors.New("ogstorage: indexer returned no roots")
	}
	// Even on a single-fragment upload roots[0] is the merkle root of
	// the file — the canonical content-addressable identifier.
	return "0g://" + roots[0].Hex(), nil
}

func (l *LiveStorage) UploadAgentCard(ctx context.Context, card AgentCard) (string, error) {
	return l.uploadJSON(ctx, card)
}

func (l *LiveStorage) AppendVoteRecord(
	ctx context.Context, _ uint64, record VoteRecord,
) (string, error) {
	return l.uploadJSON(ctx, record)
}

func (l *LiveStorage) UploadProofBundle(ctx context.Context, bundle ProofBundle) (string, error) {
	return l.uploadJSON(ctx, bundle)
}

// GetAccuracyHistory returns ErrUnimplemented for LiveStorage. The on-chain
// Registry.getAccuracy() is the canonical accuracy view; 0G Storage holds
// long-form receipts that aren't natively indexed by tokenId.
func (l *LiveStorage) GetAccuracyHistory(
	_ context.Context, _ uint64,
) ([]VoteRecord, error) {
	return nil, ErrUnimplemented
}
