package main

import (
	"net/http"

	"encoding/json"
	"strconv"

	"github.com/coreos/go-systemd/dbus"

	_ "context"

	_ "github.com/rkt/rkt/api/v1alpha"
	_ "google.golang.org/grpc"
)

/*
 * Todo: integrate with rkt API service
 * https://coreos.com/rkt/docs/latest/subcommands/api-service.html
 * https://github.com/rkt/rkt/blob/master/api/v1alpha/client_example.go
 * https://github.com/rkt/rkt/blob/master/api/v1alpha/api.proto#L460
 *
 */

/* Or: `rkt list` + `rkt cat-manifest <UUID>` + `rkt status <UUID>`
 */

type Runtime string

const (
	RunNative Runtime = "native"
	RunRkt    Runtime = "rkt"
	RunPodman Runtime = "podman"
)

type Service struct {
	Name        string
	Description string
	LoadState   string
	ActiveState string
	SubState    string
	PID         int
	Restarts    int
	Memory      uint64
	TimeChange  uint64
	Runtime     Runtime
	Container   RktInfo
}

func RegisterServiceHandlers(serveMux *http.ServeMux, c *dbus.Conn) {
	serveMux.HandleFunc("/api/services/count/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}

		svcs := GetServices(c)
		count := 0
		for _, s := range svcs {
			if s.LoadState == "not-found" {
				continue
			}
			count++
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(strconv.Itoa(count)))
	})

	serveMux.HandleFunc("/api/services/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}

		js, err := json.Marshal(GetServices(c))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})
}

func GetServices(c *dbus.Conn) []Service {
	var list []Service

	var services = []string{
		"acserver.service",
		"aimrip.service",
		"arch-repo.service",
		"bookstack.service",
		"vpn-out.service",
		"transmission.service",
		"jackett.service",
		"sonarr.service",
		"radarr.service",
		"jellyfin.service",
		"git-host.service",
		"monica.service",
		"traefik.service",
		"organizr.service",
		"quickscan.service",
		"sshd.service",
		"avahi-daemon.service",
		"org.cups.cupsd.service",
		"rkt-api.service",
		"netatalk.service",
		"nfs-server.service",
	}

	units, err := c.ListUnitsByNames(services)
	if err != nil {
		panic(err)
	}
	for _, u := range units {

		propPID, _ := c.GetServiceProperty(u.Name, "MainPID") //or ExecMainPID
		pid := int(propPID.Value.Value().(uint32))

		propRestarts, _ := c.GetServiceProperty(u.Name, "NRestarts")
		nRestarts := int(propRestarts.Value.Value().(uint32))

		propMem, _ := c.GetServiceProperty(u.Name, "MemoryCurrent")
		mem := propMem.Value.Value().(uint64)
		if mem > 1<<40 { //arbitrarily large to filter huge data
			mem = 0
		}

		timeProp, _ := c.GetUnitProperty(u.Name, "StateChangeTimestamp") // in microseconds
		lastChange := timeProp.Value.Value().(uint64)

		s := Service{
			Name:        u.Name,
			Description: u.Description,
			LoadState:   u.LoadState,
			ActiveState: u.ActiveState,
			SubState:    u.SubState,
			PID:         pid,
			Restarts:    nRestarts,
			Memory:      mem,
			TimeChange:  lastChange,
		}
		switch {
		case isPodmanService(pid):
			s.Runtime = RunPodman
		case isRktService(pid):
			s.Runtime = RunRkt
			getRktInfo(&s)
		default:
			s.Runtime = RunNative
		}

		list = append(list, s)
	}
	return list
}
