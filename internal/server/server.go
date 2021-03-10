package server

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	badger "github.com/dgraph-io/badger/v2"
	"github.com/go-chi/chi"
	"github.com/pzl/gm/internal/podman"
	"github.com/sirupsen/logrus"
)

type Server struct {
	Log          *logrus.Logger
	Router       *chi.Mux
	AssetHandler StaticHandler
	HTTP         *http.Server
	DB           *badger.DB
	DBus         *dbus.Conn
	Pod          podman.Conn
}

func New(log *logrus.Logger, port int, sh StaticHandler, db *badger.DB, dc *dbus.Conn, pc podman.Conn, devMode bool) (*Server, error) {
	router := chi.NewRouter()

	s := &Server{
		Log:          log,
		Router:       router,
		AssetHandler: sh,
		DB:           db,
		DBus:         dc,
		Pod:          pc,
		HTTP: &http.Server{
			Addr:           ":" + strconv.Itoa(port),
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   60 * time.Second,
			IdleTimeout:    300 * time.Second,
			MaxHeaderBytes: 1 << 16,
			// TLSConfig: tlsConfig,
			Handler: router,
		},
	}

	// turn off timeouts in dev mode
	// idle failure is probably interfering with hot reload
	if devMode {
		s.HTTP.ReadTimeout = 0
		s.HTTP.ReadHeaderTimeout = 0
		s.HTTP.WriteTimeout = 0
		s.HTTP.IdleTimeout = 0
	}

	s.SetupRoutes()

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	errs := make(chan error)

	tps := []string{"tcp4", "tcp6"}
	for _, l := range tps {
		s.Log.WithField("transport", l).WithField("addr", s.HTTP.Addr).Debug("opening socket")
		n, err := net.Listen(l, s.HTTP.Addr)
		if err != nil {
			return err
		}
		go func() {
			errs <- s.HTTP.Serve(n)
		}()
	}
	s.Log.Info("listening on -> " + s.HTTP.Addr)

	var err error
	select {
	case err = <-errs:
	case <-ctx.Done():
	}

	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.Log.Info("gracefully shutting down server")
	return s.HTTP.Shutdown(ctx)
}

func (s *Server) makeOutgoingClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 20,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: time.Second * 10,
			}).Dial,
			TLSHandshakeTimeout: time.Second * 10,
		},
	}
}
