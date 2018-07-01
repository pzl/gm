package main

import (
	"fmt"

	//"net/http"
	//"encoding/json"
	//"strconv"

	//"github.com/coreos/go-systemd/dbus"

	"context"
	"github.com/rkt/rkt/api/v1alpha"
	"google.golang.org/grpc"
)

/*
func main() {
	connect()
}
*/

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
