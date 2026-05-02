package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hpsing/aegis/internal/uiserver"
)

//go:embed all:web-dist
var spaFS embed.FS

func main() {
	addr := flag.String("addr", ":3000", "HTTP listen address")
	noEmbed := flag.Bool("no-embed", false, "Disable SPA serving (use with vite dev)")
	rpc := flag.String("rpc", envOr("AEGIS_RPC", "https://evmrpc-testnet.0g.ai"), "RPC for chain reads")
	aegisAddr := flag.String("aegis", envOr("AEGIS_CONTRACT", "0xa89833fBD1844763cc77C0a3aFaE32697A2F990f"), "AegisContract address")
	registryAddr := flag.String("registry", envOr("REGISTRY_CONTRACT", "0x50ce23AE35bbe43fFAd0B36FD3F567560b8EfB18"), "VerifierRegistry address")
	usdcAddr := flag.String("usdc", envOr("USDC_CONTRACT", "0xe9dA98EB0AF68cC48be7F71C29A7Bc5bA7fB45Eb"), "MockUSDC address (for in-process post-job)")
	axlList := flag.String("axl", envOr("UI_AXL_DAEMONS", "pub=http://127.0.0.1:9002,v1=http://127.0.0.1:9012,v2=http://127.0.0.1:9022,v3=http://127.0.0.1:9032"), "comma-separated role=url pairs")
	logsRoot := flag.String("logs", envOr("UI_LOGS_DIR", "logs"), "directory holding e2e-* run logs (auto-discovers latest subdir)")
	backfill := flag.Uint64("backfill", 50000, "Blocks to scan back from head for past job events on startup (0 = head only)")
	treasuryPK := flag.String("treasury-pk", envOr("UI_TREASURY_PK", os.Getenv("TREASURY_PK")), "Treasury private key (for POST /api/post-job)")
	executorPK := flag.String("executor-pk", envOr("UI_EXECUTOR_PK", os.Getenv("EXECUTOR_PK")), "Executor private key (for POST /api/post-job)")
	flag.Parse()

	hub := uiserver.NewHub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	liveCfg := uiserver.LiveConfig{
		RPC:            *rpc,
		AegisAddr:      common.HexToAddress(*aegisAddr),
		RegistryAddr:   common.HexToAddress(*registryAddr),
		USDCAddr:       common.HexToAddress(*usdcAddr),
		AXLDaemons:     parseAXLList(*axlList),
		LogsRoot:       *logsRoot,
		BackfillBlocks: *backfill,
		TreasuryPKHex:  *treasuryPK,
		ExecutorPKHex:  *executorPK,
	}
	ls, err := uiserver.NewLiveSource(ctx, hub, liveCfg)
	if err != nil {
		log.Fatalf("[ui] live source: %v", err)
	}
	if err := ls.Start(ctx); err != nil {
		log.Fatalf("[ui] live source start: %v", err)
	}
	log.Printf("[ui] live: rpc=%s aegis=%s registry=%s axl=%d nodes logs=%s",
		*rpc, liveCfg.AegisAddr.Hex()[:10]+"…", liveCfg.RegistryAddr.Hex()[:10]+"…",
		len(liveCfg.AXLDaemons), liveCfg.LogsRoot)

	cfg := uiserver.Config{
		Addr:   *addr,
		Hub:    hub,
		Source: ls,
	}

	if !*noEmbed {
		sub, err := fs.Sub(spaFS, "web-dist")
		if err != nil {
			log.Fatalf("[ui] embed sub: %v", err)
		}
		cfg.WebFS = sub
	} else {
		log.Printf("[ui] --no-embed: SPA must be served by `vite dev` on :5173")
	}

	srv := &http.Server{
		Addr:              *addr,
		Handler:           uiserver.New(cfg),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("[ui] listening on %s", *addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ui] listen: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Printf("[ui] shutdown")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		fmt.Fprintf(os.Stderr, "[ui] shutdown: %v\n", err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseAXLList(s string) map[string]string {
	out := map[string]string{}
	if s == "" {
		return out
	}
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		eq := strings.IndexByte(pair, '=')
		if eq <= 0 {
			continue
		}
		role := strings.TrimSpace(pair[:eq])
		url := strings.TrimSpace(pair[eq+1:])
		if role != "" && url != "" {
			out[role] = url
		}
	}
	return out
}
