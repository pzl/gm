package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

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

func isPodmanService(pid int) bool {
	exe, err := readCmdLine(pid)
	if err != nil {
		return false
	}
	return exe[0] == "/usr/bin/podman"
}

type podData struct{ d []byte }

func (p podData) MarshalJSON() ([]byte, error) { return p.d, nil }

func getPodmanInfo(s *Service, stats []PodmanStats) {
	exe, err := readCmdLine(s.PID)
	if err != nil {
		fmt.Printf("got error reading cmdline: %v\n", err)
		return
	}
	var name string
	for i := range exe {
		if exe[i] == "--name" {
			name = exe[i+1]
			break
		}
	}
	if name == "" {
		fmt.Printf("did not locate image name\n")
		return
	}

	out, err := exec.Command("podman", "container", "inspect", name).Output()
	if err != nil {
		fmt.Printf("error running podman inspect: %v\n", err)
		return
	}

	data := bytes.TrimSpace(out)
	data = data[1 : len(data)-1] // remove surrounding [ ]

	for _, stat := range stats {
		if stat.Name == name {
			mem := strings.TrimSpace(strings.Split(stat.MemUsage, "/")[0])
			unit := mem[len(mem)-2:]
			num, err := strconv.ParseFloat(mem[:len(mem)-2], 64)
			if err != nil {
				fmt.Printf("error parsing memory number, %s: %v\n", mem, err)
				return
			}

			switch unit {
			case "GB":
				num *= 1024
				fallthrough
			case "MB":
				num *= 1024
				fallthrough
			case "KB":
				num *= 1024
			}

			s.Memory = uint64(num)
			s.NetIO = stat.NetIO
			s.BlockIO = stat.BlockIO

			if p, err := strconv.Atoi(stat.PIDs); err == nil {
				s.PIDs = p
			}

			break
		}
	}

	s.Container = podData{d: data}
}

type PodmanStats struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	CPU      string `json:"cpu_percent"`
	MemUsage string `json:"mem_usage"`
	MemPerc  string `json:"mem_percent"`
	NetIO    string `json:"netio"`
	BlockIO  string `json:"blocki"`
	PIDs     string `json:"pids"`
}

func getPodmanStats() ([]PodmanStats, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, "podman", "container", "stats", "--no-stream", "--no-reset", "--format", "json", "-a")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)
	stats := make([]PodmanStats, 0, 20)
	if err := json.NewDecoder(out).Decode(&stats); err != nil {
		return nil, err
	}

	return stats, nil
}
