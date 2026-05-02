package uiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	zg_common "github.com/0gfoundation/0g-storage-client/common"
	"github.com/0gfoundation/0g-storage-client/indexer"
	"github.com/sirupsen/logrus"
)

// indexerURL — the 0G Storage indexer the verifiers upload through.
// Override via UI_OG_INDEXER if you point at a different testnet.
const defaultIndexerURL = "https://indexer-storage-testnet-turbo.0g.ai"

var (
	indexerOnce   sync.Once
	indexerClient *indexer.Client
	indexerErr    error
)

func getIndexerClient() (*indexer.Client, error) {
	indexerOnce.Do(func() {
		url := os.Getenv("UI_OG_INDEXER")
		if url == "" {
			url = defaultIndexerURL
		}
		// Suppress SDK chatter — see internal/ogstorage/og_service.go for
		// why a non-zero log level is required (the SDK panics on
		// PanicLevel logs from its retry path).
		l := logrus.New()
		l.SetOutput(io.Discard)
		l.SetLevel(logrus.WarnLevel)
		c, err := indexer.NewClient(url, indexer.IndexerClientOption{
			LogOption: zg_common.LogOption{Logger: l},
		})
		if err != nil {
			indexerErr = fmt.Errorf("dial indexer: %w", err)
			return
		}
		indexerClient = c
	})
	return indexerClient, indexerErr
}

// downloadProofBundle fetches a JSON file from 0G Storage by merkle
// root and returns its bytes. We download to a temp file (the SDK
// requires a filename), then read it back and clean up.
func downloadProofBundle(ctx context.Context, root string) (json.RawMessage, error) {
	if !strings.HasPrefix(root, "0x") {
		root = "0x" + root
	}
	client, err := getIndexerClient()
	if err != nil {
		return nil, err
	}

	dctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	dl, err := client.NewDownloaderFromIndexerNodes(dctx, root)
	if err != nil {
		return nil, fmt.Errorf("0g downloader: %w", err)
	}

	// SDK requires the target path NOT to exist before download.
	// CreateTemp returns the path open + empty; we reserve the unique
	// name then unlink immediately, leaving a free slot for the SDK.
	tmp, err := os.CreateTemp("", "proofbundle-*.json")
	if err != nil {
		return nil, fmt.Errorf("temp file: %w", err)
	}
	tmpName := tmp.Name()
	_ = tmp.Close()
	_ = os.Remove(tmpName)
	defer os.Remove(tmpName)

	if err := dl.Download(dctx, root, tmpName, false /* no merkle proof */); err != nil {
		return nil, fmt.Errorf("0g download %s: %w", root[:10]+"…", err)
	}
	body, err := os.ReadFile(tmpName)
	if err != nil {
		return nil, fmt.Errorf("read downloaded: %w", err)
	}
	if len(body) == 0 {
		return nil, errors.New("empty download")
	}
	// Sanity: should be JSON. If not, return as-is so the UI can show
	// the raw bytes.
	var probe any
	if json.Unmarshal(body, &probe) == nil {
		return json.RawMessage(body), nil
	}
	return json.RawMessage(body), nil
}
