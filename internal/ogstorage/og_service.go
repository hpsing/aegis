package ogstorage

import "context"

type LiveStorage struct {
	// TODO: SDK client handle, signer, node URLs.
}

func NewLiveStorage() *LiveStorage {
	return &LiveStorage{}
}

func (l *LiveStorage) UploadAgentCard(ctx context.Context, card AgentCard) (string, error) {
	return "", ErrUnimplemented
}

func (l *LiveStorage) AppendVoteRecord(
	ctx context.Context, verifierTokenID uint64, record VoteRecord,
) (string, error) {
	return "", ErrUnimplemented
}

func (l *LiveStorage) GetAccuracyHistory(
	ctx context.Context, verifierTokenID uint64,
) ([]VoteRecord, error) {
	return nil, ErrUnimplemented
}

func (l *LiveStorage) UploadProofBundle(ctx context.Context, bundle ProofBundle) (string, error) {
	return "", ErrUnimplemented
}
