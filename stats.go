package main

import (
	"net/http"

	"encoding/json"
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

func RegisterStatsHandlers(serveMux *http.ServeMux) {

	serveMux.HandleFunc("/api/stats/disk/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}
		var stats DiskStats

		blocks := DiskInfo()
		stats.Disks = blocks
		/* Old method -- only got mounted things, not unmounted blocks
		mt, err := GetMounts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, m := range mt {
			d, err := DiskUsage(m.Mount)
			if err != nil {
				continue
			}
			stats.Disks = append(stats.Disks, DiskStat{
				FS:      m.FS,
				Mount:   m.Mount,
				Type:    m.Type,
				All:     d.All,
				Used:    d.Used,
				Free:    d.Free,
				TInodes: d.TInodes,
				FInodes: d.FInodes,
			})
		}
		*/

		js, err := json.Marshal(stats)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})

}
