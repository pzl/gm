package main

import (
	//"fmt"
	"net/http"
	"time"
	"log"
	"github.com/coreos/go-systemd/dbus"

	//"strconv"
	//"strings"
	//"encoding/json"

	// dev only
	"os"
	"net/http/httputil"
	"net/url"
)

func main() {
	c, err := dbus.New()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	serveMux := http.NewServeMux()

	RegisterServiceHandlers(serveMux, c)
	RegisterSystemHandlers(serveMux, c)
	RegisterStatsHandlers(serveMux)

	//when in dev mode:
	if len(os.Args) < 2 {
		devServer, _ := url.Parse("http://localhost:3000")
		proxy := httputil.NewSingleHostReverseProxy(devServer)
		serveMux.HandleFunc("/", proxy.ServeHTTP)
	} else {
		//when in prod mode:
		serveMux.Handle("/", http.FileServer(http.Dir("frontend/dist")))
	}

	srv := &http.Server{
		Addr: ":8080",
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
		MaxHeaderBytes: 1<<16,
		//TLSConfig: tlsConfig,
		Handler: serveMux,
	}

	log.Println(srv.ListenAndServe())
}

