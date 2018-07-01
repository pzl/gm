package main

import (
	"fmt"
	//"path/filepath"
	"os"
	"bufio"
	"strconv"

	"context"
	"github.com/rkt/rkt/api/v1alpha"
	"google.golang.org/grpc"
)

func connect() error {
	conn, err := grpc.Dial("closet:15442", grpc.WithInsecure()) // move back to localhost:15441 when done remote testing
	if err != nil {
		return err
	}
	defer conn.Close()

	c := v1alpha.NewPublicAPIClient(conn)

	//list pods
	pods, err := c.ListPods(context.Background(), &v1alpha.ListPodsRequest{
		Detail:  true,
		Filters: []*v1alpha.PodFilter{},
	})
	if err != nil {
		return err
	}
	for _, p := range pods.Pods {
		fmt.Printf("POD %+v\n", p)
	}

	imgs, err := c.ListImages(context.Background(), &v1alpha.ListImagesRequest{
		Detail:  true,
		Filters: []*v1alpha.ImageFilter{},
	})
	for _, img := range imgs.Images {
		fmt.Printf("image: %q\n", img.Name)
	}

	return nil
}

func isRktService(pid int) bool {
	/* can't do, owned by root and not readable
	sym, err := filepath.EvalSymlinks("/proc/"+strconv.Itoa(pid)+"/exe")
	if err != nil {
		return false
	}
	return sym == "/usr/bin/systemd-nspawn"
	*/
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