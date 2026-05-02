package uiserver

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"
)

// Config bundles the flags / env the parent passes in.
type Config struct {
	Addr   string // e.g. ":3000"
	WebFS  fs.FS  // embedded SPA, may be nil during local dev (vite serves)
	Hub    *Hub   // SSE event hub
	Source DataSource
}

// DataSource is the read-only port the HTTP layer pulls from. Phase 0
// has a fixture implementation; phase 1 replaces it with a live source
// backed by the chain bindings + AXL clients + KH dashboard scraping.
type DataSource interface {
	State() (StateSnapshot, error)
	Jobs(limit, offset int) ([]JobSummary, int, error)
	Job(id string) (JobDetail, error)
	Verifiers() ([]VerifierInfo, error)
	Topology() (TopologyState, error)
	ProofBundle(root string) (json.RawMessage, error)
	PostJobPreview() (PostJobPreview, error)
	PostJob() (PostJobResult, error)
}

// New builds the http.Handler for the configured server.
func New(cfg Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		snap, err := cfg.Source.State()
		writeJSON(w, snap, err)
	})

	mux.HandleFunc("/api/jobs", func(w http.ResponseWriter, r *http.Request) {
		limit := atoiDefault(r.URL.Query().Get("limit"), 20)
		offset := atoiDefault(r.URL.Query().Get("offset"), 0)
		items, total, err := cfg.Source.Jobs(limit, offset)
		writeJSON(w, struct {
			Items []JobSummary `json:"items"`
			Total int          `json:"total"`
		}{items, total}, err)
	})

	mux.HandleFunc("/api/jobs/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
		if id == "" {
			http.Error(w, "missing job id", http.StatusBadRequest)
			return
		}
		job, err := cfg.Source.Job(id)
		writeJSON(w, job, err)
	})

	mux.HandleFunc("/api/verifiers", func(w http.ResponseWriter, r *http.Request) {
		vs, err := cfg.Source.Verifiers()
		writeJSON(w, vs, err)
	})

	mux.HandleFunc("/api/topology", func(w http.ResponseWriter, r *http.Request) {
		t, err := cfg.Source.Topology()
		writeJSON(w, t, err)
	})

	mux.HandleFunc("/api/proofbundle/", func(w http.ResponseWriter, r *http.Request) {
		root := strings.TrimPrefix(r.URL.Path, "/api/proofbundle/")
		if root == "" {
			http.Error(w, "missing root", http.StatusBadRequest)
			return
		}
		body, err := cfg.Source.ProofBundle(root)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})

	mux.HandleFunc("/api/post-job/preview", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "GET only", http.StatusMethodNotAllowed)
			return
		}
		preview, err := cfg.Source.PostJobPreview()
		writeJSON(w, preview, err)
	})

	mux.HandleFunc("/api/post-job", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		res, err := cfg.Source.PostJob()
		writeJSON(w, res, err)
	})

	mux.HandleFunc("/api/events", cfg.Hub.ServeHTTP)

	// SPA fallback. The embed FS is rooted at web/dist; deeper paths
	// that aren't asset hits should serve index.html so client-side
	// routing works on hard refresh.
	if cfg.WebFS != nil {
		fileServer := http.FileServer(http.FS(cfg.WebFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				fileServer.ServeHTTP(w, r)
				return
			}
			if _, err := fs.Stat(cfg.WebFS, path); err != nil {
				// SPA route — rewrite to / and let FileServer pick up index.html
				r2 := r.Clone(r.Context())
				r2.URL.Path = "/"
				fileServer.ServeHTTP(w, r2)
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	return loggingMW(mux)
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return def
	}
	return n
}

func writeJSON(w http.ResponseWriter, body any, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func loggingMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		// SSE connections stay open — log only on close.
		if r.Header.Get("Accept") != "text/event-stream" {
			fmt.Printf("[ui] %s %s %s\n", r.Method, r.URL.Path, time.Since(start))
		}
	})
}
