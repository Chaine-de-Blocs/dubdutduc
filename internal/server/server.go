package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Chaine-de-Blocs/dubdutduc/internal/chain"
	"github.com/Chaine-de-Blocs/dubdutduc/internal/music"
	"github.com/Chaine-de-Blocs/dubdutduc/internal/store"
)

type Server struct {
	Store  *store.Store
	Poller *chain.Poller
	Web    fs.FS
	Log    *slog.Logger
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/blocks", s.blocks)
	mux.HandleFunc("GET /api/blocks/{height}/score", s.score)
	mux.HandleFunc("GET /api/blocks/{height}/midi", s.midi)
	mux.Handle("GET /", http.FileServerFS(s.Web))
	return mux
}

type blockJSON struct {
	store.Block
	Key string `json:"key"`
}

func withKey(b store.Block) blockJSON {
	return blockJSON{Block: b, Key: b.Tonic + " " + b.Mode}
}

func (s *Server) blocks(w http.ResponseWriter, r *http.Request) {
	blocks, err := s.Store.List(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	out := make([]blockJSON, len(blocks))
	for i, b := range blocks {
		out[i] = withKey(b)
	}
	writeJSON(w, map[string]any{"blocks": out, "status": s.Poller.Status()})
}

type scoreNote struct {
	Midi     int     `json:"midi"`
	Start    float64 `json:"start"`
	Duration float64 `json:"duration"`
	Velocity int     `json:"velocity"`
}

type scoreTrack struct {
	Role       string      `json:"role"`
	Instrument string      `json:"instrument"`
	Name       string      `json:"name"`
	Program    int         `json:"program"`
	Notes      []scoreNote `json:"notes"`
}

type scoreBar struct {
	Chord string `json:"chord"`
	Roman string `json:"roman"`
}

type scoreJSON struct {
	blockJSON
	Length float64      `json:"length"`
	Bars   []scoreBar   `json:"bars"`
	Tracks []scoreTrack `json:"tracks"`
}

func (s *Server) score(w http.ResponseWriter, r *http.Request) {
	b, ok := s.lookup(w, r)
	if !ok {
		return
	}
	p := music.Compose(b.Hash)
	b.BPM, b.Beats = p.BPM, p.Beats
	b.Tonic, b.Mode = p.Tonic(), p.Scale.Mode.String()
	out := scoreJSON{blockJSON: withKey(b), Length: beats(p.Length())}
	for _, bar := range p.Bars {
		out.Bars = append(out.Bars, scoreBar{Chord: bar.Chord.Name(), Roman: bar.Chord.Roman()})
	}
	for i, t := range p.Tracks() {
		st := scoreTrack{
			Role:       []string{"melody", "harmony", "bass", "drums"}[i],
			Instrument: t.Instrument.ID,
			Name:       t.Instrument.Name,
			Program:    t.Instrument.Program,
			Notes:      make([]scoreNote, len(t.Notes)),
		}
		for j, n := range t.Notes {
			st.Notes[j] = scoreNote{Midi: n.Pitch, Start: beats(n.Start), Duration: beats(n.Duration), Velocity: n.Velocity}
		}
		out.Tracks = append(out.Tracks, st)
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, out)
}

func (s *Server) midi(w http.ResponseWriter, r *http.Request) {
	b, ok := s.lookup(w, r)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "audio/midi")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="block-%d.mid"`, b.Height))
	w.Write(music.Compose(b.Hash).MIDI())
}

func (s *Server) lookup(w http.ResponseWriter, r *http.Request) (store.Block, bool) {
	height, err := strconv.ParseInt(r.PathValue("height"), 10, 64)
	if err != nil {
		http.Error(w, "invalid block height", http.StatusBadRequest)
		return store.Block{}, false
	}
	b, err := s.Store.Get(r.Context(), height)
	if errors.Is(err, store.ErrNotFound) {
		http.Error(w, "block not found", http.StatusNotFound)
		return b, false
	}
	if err != nil {
		s.fail(w, err)
		return b, false
	}
	return b, true
}

func (s *Server) fail(w http.ResponseWriter, err error) {
	s.Log.Error("request failed", "err", err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}

func beats(ticks int) float64 {
	return float64(ticks) / music.PPQ
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
