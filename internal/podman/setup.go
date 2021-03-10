package podman

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Conn struct {
	URI    *url.URL
	Client *http.Client
	ctx    context.Context
}

type connMethod string

const (
	ConnUnix connMethod = "unix"
	ConnTCP  connMethod = "tcp"
	ConnSSH  connMethod = "ssh"
)

var (
	BaseURL = &url.URL{
		Scheme: "http",
		Host:   "d",
	}
)

func Connect(ctx context.Context, path string) (Conn, error) {

	//sock_dir := os.Getenv("XDG_RUNTIME_DIR")
	//sockpath := "unix:" + sock_dir + "/podman/podman.sock"

	if path == "" {
		return Conn{}, errors.New("Empty path")
	}

	var method connMethod
	if strings.HasPrefix(path, "unix:") {
		method = ConnUnix
	} else if path[0] == '/' {
		method = ConnUnix
		path = "unix://" + path
	}

	c, err := connect(ctx, method, path)
	if err != nil {
		return c, err
	}

	return c, pingConn(c)

}

func connect(ctx context.Context, method connMethod, path string) (Conn, error) {
	switch method {
	case ConnUnix:
		u, err := url.Parse(path)
		if err != nil {
			return Conn{}, err
		}
		u.Path = u.Host + "/" + u.Path
		u.Host = ""
		return Conn{
			ctx: ctx,
			URI: u,
			Client: &http.Client{
				Transport: &http.Transport{
					DialContext: func(c context.Context, _, _ string) (net.Conn, error) {
						return (&net.Dialer{}).DialContext(c, string(ConnUnix), u.Path)
					},
					DisableCompression: true,
				},
			},
		}, nil
	default:
		return Conn{}, errors.New("unsupported connection method")
	}
}

func pingConn(c Conn) error {
	_, err := c.Request(context.Background(), http.MethodGet, "/_ping", nil, nil)
	return err
}

// HTTP method conveniences
func (c Conn) Get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, path, params, nil)
}
func (c Conn) Post(ctx context.Context, path string, params url.Values, body io.Reader) ([]byte, error) {
	return c.Request(ctx, http.MethodPost, path, params, body)
}
func (c Conn) Put(ctx context.Context, path string, params url.Values, body io.Reader) ([]byte, error) {
	return c.Request(ctx, http.MethodPut, path, params, body)
}
func (c Conn) Request(ctx context.Context, method string, path string, params url.Values, sendBody io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, BaseURL.String()+path, sendBody)
	if err != nil {
		return nil, err
	}
	if len(params) > 0 {
		req.URL.RawQuery = params.Encode()
	}

	var res *http.Response
	for i := 0; i < 3; i++ {
		res, err = c.Client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i*100) * time.Millisecond)
	}
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	if res.StatusCode/100 > 2 {
		return body, fmt.Errorf("status Code: %d", res.StatusCode)
	}

	return body, nil
}
