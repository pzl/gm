package service

import (
	"strings"

	"context"

	"github.com/rkt/rkt/api/v1alpha"
	"google.golang.org/grpc"
)

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

func getRktInfo(s *Service) {
	if sysPods == nil {
		getSystemPods()
	}

	// BUG(pzl): early-matches a possibly old/garbage container for a given svc
PodSearch:
	for _, p := range sysPods {
		for _, a := range p.Apps {
			if serviceMatchesContainer(s.Name, a.Name) {
				s.Container = ContainerInfo(p)
				break PodSearch
			}
		}
	}
}

func serviceMatchesContainer(srv string, ctr string) bool {
	if srv == ctr+".service" {
		return true
	}
	if srv == strings.Split(ctr, "-")[0]+".service" {
		return true
	}
	if srv == "entertainment.service" && (ctr == "sonarr" || ctr == "radarr" || ctr == "jackett") {
		return true
	}
	return false
}
