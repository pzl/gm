package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pzl/gm/internal/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type Config struct {
	Port    int
	DevMode bool
	DBPath  string
	PodSock string
	Handler server.StaticHandler
}

func setup() (Config, context.Context, context.CancelFunc, *logrus.Logger) {
	verbose := pflag.CountP("verbose", "v", "increased logging. Use Multiple times for more info")
	j := pflag.BoolP("json", "j", false, "output logs in JSON format")
	port := pflag.IntP("port", "p", 2556, "Listening port")
	dev := pflag.BoolP("dev", "d", false, "enable development mode. Listens to npm dev server for static assets")
	sockPath := pflag.StringP("socket", "k", "/run/podman/podman.sock", "Podman service socket path")
	dbpath := pflag.StringP("storage", "s", "db", "path to database directory")
	pflag.Lookup("storage").NoOptDefVal = ":MEMORY:"

	pflag.Parse()
	if port == nil || *port < 1 {
		*port = 2556
	}
	if dbpath == nil {
		*dbpath = "db"
	}

	log := logrus.New()
	SetLogLevel(log, *verbose)
	SetLogMode(log, *j)
	ah := setAssetHandler(*dev, log)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	return Config{
		Port:    *port,
		Handler: ah,
		DevMode: *dev,
		DBPath:  *dbpath,
		PodSock: *sockPath,
	}, ctx, cancel, log

}

type SPAFileSystem struct{ http.FileSystem }

func (s SPAFileSystem) Open(name string) (http.File, error) {
	f, err := s.FileSystem.Open(name)
	if os.IsNotExist(err) {
		//send it to index.html
		return s.FileSystem.Open("index.html")
	}
	return f, err
}

func setAssetHandler(devMode bool, log *logrus.Logger) server.StaticHandler {
	var assetHandler http.Handler
	if devMode {
		log.Info("dev mode enabled. Listening to npm dev server at localhost:3000")
		devServer, _ := url.Parse("http://localhost:3000")
		proxy := httputil.NewSingleHostReverseProxy(devServer)
		assetHandler = proxy
	} else {
		log.Info("serving in production mode, with precompiled assets")
		assetHandler = http.FileServer(SPAFileSystem{assets})
	}
	return assetHandler
}

// given -v or -vv or whatever on the command line
// this sets a logrus instance to that verbosity
func SetLogLevel(log *logrus.Logger, level int) {
	lvls := []logrus.Level{
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}
	if level > 3 {
		level = 3
	} else if level < 0 {
		level = 0
	}
	log.SetLevel(lvls[level])
}

// sets a logrus formatter to either use JSON or not
// with some opinionated formats otherwide
func SetLogMode(log *logrus.Logger, useJSON bool) {
	if useJSON {
		log.Formatter = UTCFormatter{&logrus.JSONFormatter{
			TimestampFormat: time.RFC1123,
		}}
	} else {
		log.Formatter = UTCFormatter{&logrus.TextFormatter{
			TimestampFormat:  time.RFC1123,
			FullTimestamp:    true,
			QuoteEmptyFields: true,
			ForceColors:      true,
		}}
	}
}

type UTCFormatter struct{ logrus.Formatter }

func (u UTCFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}
