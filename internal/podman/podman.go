package podman

import (
	"context"
	"net/url"
)

func (c Conn) Info(ctx context.Context) ([]byte, error) { return c.Get(ctx, "/v3/libpod/info", nil) }

func (c Conn) ContainerInfo(ctx context.Context, ctrs ...string) ([]byte, error) {
	v := url.Values{}
	for _, c := range ctrs {
		v.Add("filters", "id="+c)
	}
	return c.Get(ctx, "/v3/libpod/containers/json", v)
}

func (c Conn) Inspect(ctx context.Context, name string) ([]byte, error) {
	return c.Get(ctx, "/v3/libpod/containers/"+name+"/json", nil)
}
