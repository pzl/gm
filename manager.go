package main

import (
	//"fmt"

	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	// dev only
	"net/http/httputil"
	"net/url"
)

type ctxKey int

const (
	logKey ctxKey = iota
	dbusKey
	fileKey
)

func main() {
	verbose := pflag.CountP("verbose", "v", "increased logging. Use multiple times for more")
	j := pflag.BoolP("json", "j", false, "output logs in JSON formt")
	port := pflag.IntP("port", "p", 2556, "Listening port")
	dev := pflag.BoolP("dev", "d", false, "enable development mode. Listens to npm dev server for static assets")
	svcFile := pflag.StringP("services", "s", "/srv/apps/manager/services.txt", "Path to file containing systemD services to track, one per line")
	log := logrus.New()

	pflag.Parse()

	ctx := context.Background()
	ctx = context.WithValue(ctx, fileKey, *svcFile)

	logLvl(log, *verbose)
	if *j {
		log.Formatter = UTCFormatter{&logrus.JSONFormatter{
			TimestampFormat: time.RFC1123,
		}}
	} else {
		log.Formatter = UTCFormatter{&logrus.TextFormatter{
			TimestampFormat:  time.RFC1123,
			QuoteEmptyFields: true,
		}}
	}

	log.WithFields(logrus.Fields{
		"verbose:":     *verbose,
		"json":         *j,
		"port":         *port,
		"service file": *svcFile,
	}).Trace("parsed flags")

	ctx = context.WithValue(ctx, logKey, log)

	c, err := dbus.New()
	if err != nil {
		log.WithError(err).Error("unable to setup DBUS")
		panic(err)
	}
	defer c.Close()
	ctx = context.WithValue(ctx, dbusKey, c)

	log.Trace("connected to DBUS")

	serveMux := http.NewServeMux()

	RegisterServiceHandlers(serveMux, ctx)
	RegisterSystemHandlers(serveMux, ctx)
	RegisterStatsHandlers(serveMux, ctx)

	//when in dev mode:
	if *dev {
		log.Info("dev mode enabled. Listening to npm dev server at localhost:3000")
		devServer, _ := url.Parse("http://localhost:3000")
		proxy := httputil.NewSingleHostReverseProxy(devServer)
		serveMux.HandleFunc("/", proxy.ServeHTTP)
	} else {
		//when in prod mode:
		log.Info("serving in production mode. Fileserver at frontend/dist")
		//@todo: compile these assets in
		serveMux.Handle("/", http.FileServer(http.Dir("frontend/dist")))
	}

	srv := &http.Server{
		Addr:           ":" + strconv.Itoa(*port),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 16,
		//TLSConfig: tlsConfig,
		Handler: serveMux,
	}

	log.WithField("port", *port).Info("Listening for HTTP requests")
	err = srv.ListenAndServe()
	if err != nil {
		log.WithError(err).Error("error running ListenAndServe")
	}
	log.Info("exiting")
}

// sets logger level, based on -v count
func logLvl(log *logrus.Logger, v int) {
	lvls := []logrus.Level{
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}
	if v > 3 {
		v = 3
	} else if v < 0 {
		v = 0
	}
	log.SetLevel(lvls[v])
}

type UTCFormatter struct{ logrus.Formatter }

func (u UTCFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}
