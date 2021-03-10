package main

import (
	"context"
	"net/http"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/coreos/go-systemd/v22/util"
	badger "github.com/dgraph-io/badger/v2"
	"github.com/pzl/gm/internal/podman"
	"github.com/pzl/gm/internal/server"
)

//go:generate go run assets_gen.go

func main() {
	cfg, ctx, cancel, log := setup()
	defer cancel()

	shutDownTimer := 20 * time.Second

	dbopts := badger.DefaultOptions(cfg.DBPath).WithLogger(log)
	if cfg.DBPath == ":MEMORY:" {
		dbopts.InMemory = true
		dbopts.Dir = ""
		dbopts.ValueDir = ""
	}

	db, err := badger.Open(dbopts)
	if err != nil {
		log.WithError(err).Error("unable to open database")
		panic(err)
	}
	defer db.Close()

	dc, err := dbus.NewWithContext(ctx)
	if err != nil {
		log.WithError(err).Error("unable to setup DBUS")
		panic(err)
	}
	defer dc.Close()

	if ok := util.IsRunningSystemd(); !ok {
		panic("This tool requires Systemd to be running")
	}

	pc, err := podman.Connect(ctx, cfg.PodSock)
	if err != nil {
		log.WithError(err).Error("unable to setup podman connection")
		panic(err)
	}

	srv, err := server.New(log, cfg.Port, cfg.Handler, db, dc, pc, cfg.DevMode)
	if err != nil {
		log.WithError(err).Error("error creating server")
		panic(err)
	}

	err = srv.Start(ctx)
	defer func() {
		shutCtx, _ := context.WithTimeout(context.Background(), shutDownTimer)
		if err := srv.Shutdown(shutCtx); err != nil {
			log.WithError(err).Error("unable to gracefully shutdown")
		}
	}()
	if err != nil && err != http.ErrServerClosed {
		log.WithError(err).Error("server error")
	}
}
