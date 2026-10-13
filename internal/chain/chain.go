package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Chaine-de-Blocs/dubdutduc/internal/music"
	"github.com/Chaine-de-Blocs/dubdutduc/internal/store"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

type apiBlock struct {
	ID        string `json:"id"`
	Height    int64  `json:"height"`
	Timestamp int64  `json:"timestamp"`
}

// Blocks returns up to 15 blocks, newest first, ending at height (or at the tip when height < 0).
func (c *Client) Blocks(ctx context.Context, height int64) ([]apiBlock, error) {
	url := strings.TrimRight(c.BaseURL, "/") + "/api/v1/blocks"
	if height >= 0 {
		url += fmt.Sprintf("/%d", height)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, res.Status)
	}
	var blocks []apiBlock
	if err := json.NewDecoder(res.Body).Decode(&blocks); err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	return blocks, nil
}

type Status struct {
	Tip      int64     `json:"tip"`
	SyncedAt time.Time `json:"syncedAt"`
	Syncing  bool      `json:"syncing"`
	Error    string    `json:"error,omitempty"`
}

type Poller struct {
	Client   *Client
	Store    *store.Store
	Backfill int64
	Interval time.Duration
	Log      *slog.Logger

	mu     sync.Mutex
	status Status
}

func (p *Poller) Status() Status {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status
}

func (p *Poller) Run(ctx context.Context) {
	t := time.NewTicker(p.Interval)
	defer t.Stop()
	for {
		p.sync(ctx)
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (p *Poller) sync(ctx context.Context) {
	p.setStatus(func(s *Status) { s.Syncing = true })
	tip, err := p.fill(ctx)
	if ctx.Err() != nil {
		return
	}
	p.setStatus(func(s *Status) {
		s.Syncing = false
		if err != nil {
			s.Error = err.Error()
			return
		}
		s.Error = ""
		s.Tip = tip
		s.SyncedAt = time.Now()
	})
	if err != nil {
		p.Log.Warn("sync failed", "err", err)
	}
}

func (p *Poller) setStatus(f func(*Status)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	f(&p.status)
}

// fill stores the latest blocks, then every missing height from the last stored block
// (or the backfill window on first run) up to the tip.
func (p *Poller) fill(ctx context.Context) (int64, error) {
	maxHeight, stored, err := p.Store.MaxHeight(ctx)
	if err != nil {
		return 0, err
	}
	latest, err := p.Client.Blocks(ctx, -1)
	if err != nil {
		return 0, err
	}
	if len(latest) == 0 {
		return 0, fmt.Errorf("mempool returned no blocks")
	}
	if err := p.save(ctx, latest); err != nil {
		return 0, err
	}
	tip := latest[0].Height

	lo := tip - p.Backfill + 1
	if stored && maxHeight+1 < lo {
		lo = maxHeight + 1
	}
	have, err := p.Store.Heights(ctx, lo, tip)
	if err != nil {
		return 0, err
	}

	for h := tip; h >= lo; {
		if have[h] {
			h--
			continue
		}
		blocks, err := p.Client.Blocks(ctx, h)
		if err != nil {
			return 0, err
		}
		if len(blocks) == 0 {
			break
		}
		if err := p.save(ctx, blocks); err != nil {
			return 0, err
		}
		for _, b := range blocks {
			have[b.Height] = true
		}
		h = blocks[len(blocks)-1].Height - 1
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(300 * time.Millisecond):
		}
	}
	return tip, nil
}

func (p *Poller) save(ctx context.Context, blocks []apiBlock) error {
	for _, b := range blocks {
		piece := music.Compose(b.ID)
		added, err := p.Store.Save(ctx, store.Block{
			Height: b.Height,
			Hash:   piece.Hash,
			Time:   b.Timestamp,
			Tonic:  piece.Tonic(),
			Mode:   piece.Scale.Mode.String(),
			BPM:    piece.BPM,
			Beats:  piece.Beats,
		})
		if err != nil {
			return err
		}
		if added {
			p.Log.Info("new piece", "height", b.Height, "key", piece.Key(), "bpm", piece.BPM)
		}
	}
	return nil
}
