package main

import (
	"os"
	"bufio"
	"strconv"
	"strings"

	"context"
	"github.com/rkt/rkt/api/v1alpha"
	"google.golang.org/grpc"
)

func isRktService(pid int) bool {
	file, err := os.Open("/proc/"+strconv.Itoa(pid)+"/cmdline")
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

var sysPods []*v1alpha.Pod

func getSystemPods() {
	conn, err := grpc.Dial("localhost:15441", grpc.WithInsecure())
	if err != nil {
		return
	}
	defer conn.Close()
	c := v1alpha.NewPublicAPIClient(conn)

	pods, err := c.ListPods(context.Background(), &v1alpha.ListPodsRequest{
		Detail:  true,
		Filters: []*v1alpha.PodFilter{},
	})
	if err != nil {
		return
	}
	sysPods = pods.Pods
}

type RktInfo *v1alpha.Pod

func getRktInfo(s *Service) {
	if sysPods == nil {
		getSystemPods()
	}

	// BUG(pzl): early-matches a possibly old/garbage container for a given svc
	PodSearch:
	for _, p := range sysPods {
		for _, a := range p.Apps {
			if serviceMatchesContainer(s.Name, a.Name) {
				s.Container = RktInfo(p)
				break PodSearch
			}
		}
	}
}

func serviceMatchesContainer(srv string, ctr string) bool {
	if srv == ctr + ".service" {
		return true
	}
	if srv == strings.Split(ctr, "-")[0] + ".service" {
		return true
	}
	if srv == "entertainment.service" && (ctr == "sonarr" || ctr == "radarr" || ctr == "jackett") {
		return true
	}
	return false
}
