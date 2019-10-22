package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"encoding/json"
	"strconv"

	"github.com/coreos/go-systemd/dbus"
	"github.com/sirupsen/logrus"

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

type ContainerInfo interface{}

type Service struct {
	Name        string
	Description string
	LoadState   string
	ActiveState string
	SubState    string
	PID         int
	Restarts    int
	Memory      uint64
	NetIO       string
	BlockIO     string
	PIDs        int
	TimeChange  uint64
	Runtime     Runtime
	Container   ContainerInfo
}

func RegisterServiceHandlers(serveMux *http.ServeMux, ctx context.Context) {
	log := ctx.Value(logKey).(*logrus.Logger)

	serveMux.HandleFunc("/api/services/count/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}

		log.Debug("fetching services")
		svcs := GetServices(ctx)
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

		log.Debug("fetching services")
		svc := GetServices(ctx)
		if svc == nil {
			log.Info("got nil services")
			http.Error(w, "nil response", 500)
			return
		}
		log.Debug("encoding response as json")
		js, err := json.Marshal(svc)
		if err != nil {
			log.WithError(err).Error("error marshalling services to json")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})
}

func GetServices(ctx context.Context) []Service {
	var list []Service

	c := ctx.Value(dbusKey).(*dbus.Conn)
	f := ctx.Value(fileKey).(string)
	log := ctx.Value(logKey).(*logrus.Logger)

	data, err := ioutil.ReadFile(f)
	if err != nil {
		log.WithError(err).WithField("file", f).Error("error reading services file")
		return nil
	}

	lines := bytes.Split(data, []byte{'\n'})
	services := make([]string, len(lines))
	for i := range lines {
		services[i] = string(lines[i])
	}
	log.WithField("n lines", len(lines)).Trace("read from services file")

	units, err := c.ListUnitsByNames(services)
	if err != nil {
		log.WithError(err).Error("error listing systemd units via dbus")
		return nil
	}

	podstats, err := getPodmanStats()
	if err != nil {
		log.WithError(err).Error("error getting podman stats")
		return nil
	}

	for _, u := range units {
		log.WithField("name", u.Name).Debug("getting dbus systemd info on service")
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
			log.WithField("name", u.Name).Debug("is a podman service. Enriching with container info")
			s.Runtime = RunPodman
			getPodmanInfo(&s, podstats)
		case isRktService(pid):
			log.WithField("name", u.Name).Debug("is a rkt service. Enriching with container info")
			s.Runtime = RunRkt
			getRktInfo(&s)
		default:
			log.WithField("name", u.Name).Debug("is not a container service")
			s.Runtime = RunNative
		}

		list = append(list, s)
	}
	return list
}
