package main

import (
	"context"
	"net/http"

	"encoding/json"

	"github.com/sirupsen/logrus"
	//"strconv"
)

type DiskStats struct {
	Disks []*BlockDevice
}

/*
type DiskStat struct {
	FS      string
	Mount   string
	Type    string
	All     uint64
	Used    uint64
	Free    uint64
	TInodes uint64
	FInodes uint64
}
*/

func RegisterStatsHandlers(serveMux *http.ServeMux, ctx context.Context) {
	log := ctx.Value(logKey).(*logrus.Logger)

	serveMux.HandleFunc("/api/stats/disk/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}
		var stats DiskStats

		blocks := DiskInfo()
		stats.Disks = blocks

		js, err := json.Marshal(stats)
		if err != nil {
			log.WithError(err).Error("error json marshaling disk stat info")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})

}
