package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	// used for getting a field for sorting unparsed list responses

	"github.com/buger/jsonparser"
	"github.com/go-chi/chi"
	jsoniter "github.com/json-iterator/go"
	"github.com/pzl/gm/internal/store"
	"github.com/pzl/gm/internal/sysd"
)

var jsCfg = jsoniter.Config{
	EscapeHTML:                    true,
	SortMapKeys:                   false,
	ValidateJsonRawMessage:        false,
	MarshalFloatWith6Digits:       true,
	ObjectFieldMustBeSimpleString: false,
}.Froze()

type Svc struct {
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
}

func (s *Server) SvcHandlerGet(w http.ResponseWriter, r *http.Request) {
	unit := chi.URLParam(r, "unit")
	response, err := sysd.GetServices(r.Context(), s.DBus, unit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	svc := response[0]

	if svc.Extended != nil && svc.Extended.Runtime == "podman" {
		var name string
		for i, arg := range svc.Cmdline {
			if arg == "--name" {
				name = svc.Cmdline[i+1]
				break
			}
		}
		if name != "" {
			podinfo, err := s.Pod.Inspect(r.Context(), name)
			if err == nil {
				svc.External = podinfo
			}
		} else {
			s.Log.WithField("Cmdline", svc.Cmdline).Warn("Name of container not found")
		}
	}

	jsCfg.NewEncoder(w).Encode(svc)
}

func (s *Server) SvcHandlerSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	if len(q) < 2 {
		_, _ = w.Write([]byte("[]"))
		return
	}

	stats, err := s.DBus.ListUnitFilesByPatternsContext(r.Context(), nil, []string{"*" + q + "*"})
	if err != nil {
		s.Log.WithError(err).Error("unable to do unit typeahead")
		_, _ = w.Write([]byte("[]"))
		return
	}

	names := make([]string, len(stats))
	for i, u := range stats {
		names[i] = filepath.Base(u.Path)
	}
	jsCfg.NewEncoder(w).Encode(names)
}

func (s *Server) SvcHandlerList(w http.ResponseWriter, r *http.Request) {
	response, err := store.GetAllForType(s.DB, store.SvcKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}
	names := make([]string, len(response))
	for i, r := range response {
		unit, err := jsonparser.GetString(r, "name")
		if err == nil {
			names[i] = unit
		}
	}
	svc, err := sysd.GetServices(r.Context(), s.DBus, names...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	if r.URL.Query().Get("extra") != "" {
		for i := range svc {
			if svc[i].Extended == nil {
				continue
			}
			if svc[i].Extended.Runtime != "podman" {
				continue
			}
			var name string
			for j, arg := range svc[i].Cmdline {
				if arg == "--name" {
					name = svc[i].Cmdline[j+1]
					break
				}
			}
			if name != "" {
				podinfo, err := s.Pod.Inspect(r.Context(), name)
				if err == nil {
					svc[i].External = podinfo
				}
			} else {
				s.Log.WithField("Cmdline", svc[i].Cmdline).Warn("Name of container not found")
			}
		}
	}

	jsCfg.NewEncoder(w).Encode(svc)
}

func (s *Server) SvcHandlerCreate(w http.ResponseWriter, r *http.Request) {
	type SvcCreateRequest struct {
		// fields...
		Name string `json:"name"`
	}
	var cr SvcCreateRequest
	if err := jsCfg.NewDecoder(r.Body).Decode(&cr); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	if strings.TrimSpace(cr.Name) == "" { // validate request however
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid request"}`))
		return
	}

	// assign fields to the Svc
	svc := Svc{
		Name:      cr.Name,
		CreatedAt: time.Now().Unix(),
	}

	if err := store.WriteType(s.DB, store.SvcKey, svc.Name, svc, 0, 0); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"Error creating service: ` + err.Error() + `"}`))
		return
	}
	_ = jsCfg.NewEncoder(w).Encode(svc)

}

func (s *Server) SvcHandlerAction(w http.ResponseWriter, r *http.Request) {
	unit := chi.URLParam(r, "unit")

	req := struct {
		Action string `json:"action"`
	}{}
	if err := jsCfg.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	var do DBusAction

	switch strings.ToLower(strings.TrimSpace(req.Action)) {
	case "start":
		do = s.DBus.StartUnitContext
	case "stop":
		do = s.DBus.StopUnitContext
	case "restart":
		do = s.DBus.RestartUnitContext
	case "enable":
		install, changes, err := s.DBus.EnableUnitFilesContext(r.Context(), []string{unit}, false, true)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"Error enabling service: ` + err.Error() + `"}`))
			return
		}
		jsCfg.NewEncoder(w).Encode(map[string]interface{}{
			"install": install,
			"changes": changes,
		})
		return
	case "disable":
		changes, err := s.DBus.DisableUnitFilesContext(r.Context(), []string{unit}, false)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"Error disabling service: ` + err.Error() + `"}`))
			return
		}
		jsCfg.NewEncoder(w).Encode(map[string]interface{}{
			"changes": changes,
		})
		return
	case "update":
		// custom podman container stuff
	default:
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"unknown action: ` + req.Action + `"}`))
		return
	}

	if do == nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"unknown action: ` + req.Action + `"}`))
		return
	}

	if err := doDbusAction(r.Context(), do, unit, "replace"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}
	_, _ = w.Write([]byte(`{"message":"success"}`))

}

type DBusAction func(context.Context, string, string, chan<- string) (int, error)

func doDbusAction(ctx context.Context, do DBusAction, unit string, mode string) error {
	done := make(chan string)
	_, err := do(ctx, unit, mode, done) // first param is a JobID
	if err != nil {
		return err
	}
	select {
	case result := <-done:
		if result == "done" {
			return nil
		}
		return fmt.Errorf("job %s", result)
	case <-time.After(30 * time.Second):
		go func() {
			<-done // make sure the write doesn't block
		}()
		return errors.New("timed out")
	}
}

func (s *Server) SvcHandlerDelete(w http.ResponseWriter, r *http.Request) {
	unit := chi.URLParam(r, "unit")

	if err := store.DeleteRecord(s.DB, store.SvcKey, unit); err != nil {
		s.Log.WithError(err).Error("error deleting thing")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{ "status":"error", "error": "` + err.Error() + `"}`))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
