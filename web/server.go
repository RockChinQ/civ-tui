package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"

	"github.com/RockChinQ/civ-tui/game"
)

//go:embed static
var staticFS embed.FS

// Server holds the single active game and serves both the JSON API and static
// front-end assets. It is intended for single-player local use, so all access
// is serialized through a single mutex.
type Server struct {
	mu   sync.Mutex
	game *game.Game
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) withGame(w http.ResponseWriter, fn func(g *game.Game) any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.game == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no active game"})
		return
	}
	resp := fn(s.game)
	if resp == nil {
		resp = s.gameStateLocked()
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()

	// Static assets
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}
	mux.Handle("/", http.FileServer(http.FS(sub)))

	// Game state
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/new", s.handleNew)
	mux.HandleFunc("/api/saves", s.handleSaves)
	mux.HandleFunc("/api/save", s.handleSave)
	mux.HandleFunc("/api/load", s.handleLoad)

	// Actions
	mux.HandleFunc("/api/move", s.handleMove)
	mux.HandleFunc("/api/found-city", s.handleFoundCity)
	mux.HandleFunc("/api/wait", s.handleWait)
	mux.HandleFunc("/api/end-turn", s.handleEndTurn)
	mux.HandleFunc("/api/ranged", s.handleRanged)
	mux.HandleFunc("/api/improvement", s.handleImprovement)
	mux.HandleFunc("/api/build", s.handleBuild)
	mux.HandleFunc("/api/research", s.handleResearch)
	mux.HandleFunc("/api/set-dest", s.handleSetDest)
	mux.HandleFunc("/api/clear-dest", s.handleClearDest)

	log.Printf("civ-web listening on http://%s", addr)
	return http.ListenAndServe(addr, mux)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		log.Printf("encode error: %v", err)
	}
}

func decodeJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return fmt.Errorf("empty body")
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
