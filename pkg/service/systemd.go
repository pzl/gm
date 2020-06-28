package service

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"os"

	"encoding/json"
	"strconv"

	"github.com/coreos/go-systemd/dbus"
	"github.com/pzl/manager/pkg/config"
	"github.com/sirupsen/logrus"

	_ "context"

	_ "github.com/rkt/rkt/api/v1alpha"
	_ "google.golang.org/grpc"
)

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
	log := ctx.Value(config.LogKey).(*logrus.Logger)

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

	c := ctx.Value(config.DbusKey).(*dbus.Conn)
	f := ctx.Value(config.FileKey).(string)
	log := ctx.Value(config.LogKey).(*logrus.Logger)

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
		case isPodman(pid):
			log.WithField("name", u.Name).Debug("is a podman service. Enriching with container info")
			s.Runtime = RunPodman
			getPodmanInfo(&s, podstats)
		case isRkt(pid):
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

func isRkt(pid int) bool {
	file, err := os.Open("/proc/" + strconv.Itoa(pid) + "/cmdline")
	if err != nil {
		return false
	}
	defer file.Close()
	onNul := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i := 0; i < len(data); i++ {
			if data[i] == '\x00' {
				return i + 1, data[:i], nil
			}
		}
		return 0, data, bufio.ErrFinalToken
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(onNul)
	scanner.Scan()
	exe := scanner.Text()
	return exe == "/usr/bin/systemd-nspawn"
}

func isPodman(pid int) bool {
	exe, err := readCmdLine(pid)
	if err != nil {
		return false
	}
	return exe[0] == "/usr/bin/podman"
}
