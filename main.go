package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Chaine-de-Blocs/dubdutduc/internal/chain"
	"github.com/Chaine-de-Blocs/dubdutduc/internal/server"
	"github.com/Chaine-de-Blocs/dubdutduc/internal/store"
)

//go:embed web
var web embed.FS

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "data/blocks.db", "SQLite catalog path")
	mempool := flag.String("mempool", "https://mempool.space", "mempool.space compatible API")
	backfill := flag.Int64("backfill", 100, "recent blocks to load on first run")
	interval := flag.Duration("interval", 15*time.Second, "tip polling interval")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(log, *addr, *dbPath, *mempool, *backfill, *interval); err != nil {
		log.Error("exit", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger, addr, dbPath, mempool string, backfill int64, interval time.Duration) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer st.Close()

	poller := &chain.Poller{
		Client:   &chain.Client{BaseURL: mempool, HTTP: &http.Client{Timeout: 20 * time.Second}},
		Store:    st,
		Backfill: backfill,
		Interval: interval,
		Log:      log,
	}
	go poller.Run(ctx)

	static, err := fs.Sub(web, "web")
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           (&server.Server{Store: st, Poller: poller, Web: static, Log: log}).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdown)
	}()

	log.Info("listening", "addr", addr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
