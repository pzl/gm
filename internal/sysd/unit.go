package sysd

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"
	jsoniter "github.com/json-iterator/go"
)

type Service struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	LoadState   string              `json:"load_state"`
	ActiveState string              `json:"active_state"`
	SubState    string              `json:"sub_state"`
	Followed    string              `json:"followed"`
	Path        string              `json:"path"`
	JobID       uint32              `json:"job_id"`
	JobType     string              `json:"job_type"`
	JobPath     string              `json:"job_path"`
	Cmdline     []string            `json:"cmdline"`
	FileState   string              `json:"file_state"`
	Extended    *Extended           `json:"extended"`
	External    jsoniter.RawMessage `json:"external,omitempty"`
}

type Extended struct {
	PID        int    `json:"pid"`
	Restarts   int    `json:"restarts"`
	Memory     uint64 `json:"mem"`
	LastChange uint64 `json:"last_change"` // in microsecondss
	Runtime    string `json:"runtime"`
}

var NotFoundError = errors.New("Unit not found")

func GetServices(ctx context.Context, d *dbus.Conn, units ...string) ([]Service, error) {
	result, err := d.ListUnitsByNamesContext(ctx, units)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, NotFoundError
	}

	fileStates := make(map[string]string, len(result))
	if uf, err := d.ListUnitFilesByPatternsContext(ctx, nil, units); err == nil {
		for _, f := range uf {
			fileStates[filepath.Base(f.Path)] = f.Type
		}
	}

	svc := make([]Service, len(result))

	for i, r := range result {
		s := Service{
			Name:        r.Name,
			Description: r.Description,
			LoadState:   r.LoadState,
			ActiveState: r.ActiveState,
			SubState:    r.SubState,
			Followed:    r.Followed,
			Path:        string(r.Path),
			JobID:       r.JobId,
			JobType:     r.JobType,
			JobPath:     string(r.JobPath),
		}
		if fs, ok := fileStates[r.Name]; ok {
			s.FileState = fs
		}
		EnrichExtended(ctx, d, &s)
		svc[i] = s
	}

	return svc, nil
}

func EnrichExtended(ctx context.Context, d *dbus.Conn, s *Service) {
	if s.Extended == nil {
		s.Extended = &Extended{PID: -1, Restarts: -1}
	}

	if propPID, err := d.GetServicePropertyContext(ctx, s.Name, "MainPID"); err == nil {
		if n, ok := propPID.Value.Value().(uint32); ok {
			s.Extended.PID = int(n)
		}
	}

	if propRestarts, err := d.GetServicePropertyContext(ctx, s.Name, "NRestarts"); err == nil {
		if n, ok := propRestarts.Value.Value().(uint32); ok {
			s.Extended.Restarts = int(n)
		}
	}

	if propMem, err := d.GetServicePropertyContext(ctx, s.Name, "MemoryCurrent"); err == nil {
		if n, ok := propMem.Value.Value().(uint64); ok {
			if n > 1<<40 { // filter out arbitrarily large numbers
				n = 0
			}
			s.Extended.Memory = n
		}
	}
	if timeProp, err := d.GetUnitPropertyContext(ctx, s.Name, "StateChangeTimestamp"); err == nil {
		if n, ok := timeProp.Value.Value().(uint64); ok {
			s.Extended.LastChange = n
		}
	}

	if cmds, err := readCmdLine(s.Extended.PID); err == nil {
		s.Cmdline = cmds
	}

	switch {
	case isPodman(s):
		s.Extended.Runtime = "podman"
	default:
		s.Extended.Runtime = "native"
	}
}

func isPodman(s *Service) bool {
	return len(s.Cmdline) > 0 && strings.HasSuffix(s.Cmdline[0], "podman")
}
func readCmdLine(pid int) ([]string, error) {
	data, err := ioutil.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	if err != nil {
		return nil, err
	}
	b := bytes.Split(data, []byte{0})
	s := make([]string, len(b))
	for i := range b {
		s[i] = string(b[i])
	}
	return s, nil
}
