package server

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pzl/mstk"
	"github.com/pzl/mstk/logger"
)

type StaticHandler http.Handler

func (s *Server) SetupRoutes() {
	s.Router.Use(middleware.RealIP)
	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.RequestLogger(logger.NewChi(s.Log)))
	s.Router.Use(middleware.Heartbeat("/ping"))
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(cors)

	s.routeAPI()
	s.routeWeb()
}

func (s *Server) routeWeb() {
	s.Router.Get("/_nuxt/*", s.AssetHandler.ServeHTTP)
	files := []string{"favicon.ico"} // any other root-level static things
	for _, f := range files {
		s.Router.Get("/"+f, s.AssetHandler.ServeHTTP)
	}

	s.Router.Get("/*", s.AssetHandler.ServeHTTP)
}

func (s *Server) routeAPI() {
	s.Router.Route("/api/v1", func(v1 chi.Router) {
		v1.Use(contentJSON)
		v1.Use(mstk.APIVer(1))

		v1.Get("/stats/disk", nil)

		v1.Get("/services", s.SvcHandlerList)
		v1.Get("/services/search", s.SvcHandlerSearch)    // query for a unit name
		v1.Post("/services", s.SvcHandlerCreate)          // add a new service to be watched
		v1.Get("/services/{unit}", s.SvcHandlerGet)       // get extended info for a unit
		v1.Post("/services/{unit}", s.SvcHandlerAction)   // Send a command. Reload, Restart, Kill, etc the unit
		v1.Delete("/services/{unit}", s.SvcHandlerDelete) // removes unit from being watched

		v1.Post("/system/reload", s.SystemReloadSysD) // systemctl daemon-reload
		v1.Get("/system/disk", s.SystemGetDisk)       // HDD , block info
		v1.Get("/system/versions", s.SystemGetVersions)
		//v1.Get("/system/vpn", nil) <-- get /services/vpn-out.service and check status
		v1.Get("/system/memory", s.SystemGetMemory)

		v1.Get("/ws", s.WebsocketHandler)

	})
}

func cors(next http.Handler) http.Handler {

	allowedHeaders := strings.Join([]string{
		"Origin",
		"x-Requested-With",
		"Content-Type",
		"Accept",
		"User",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Range",
		"Accept-Ranges",
	}, ", ")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			// OPTIONS -- let's handle fully in here

			w.Header().Add("Vary", "Origin")
			w.Header().Add("Vary", "Access-Control-Request-Method")
			w.Header().Add("Vary", "Access-Control-Request-Headers")

			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", strings.ToUpper(r.Header.Get("Access-Control-Request-Method")))
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			w.WriteHeader(http.StatusOK)
			return
			// and be done. DO NOT serveHTTP
		}
		// NOT 'OPTIONS', so do some header altering and pass on

		w.Header().Add("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		next.ServeHTTP(w, r)
	})
}

func contentJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}
