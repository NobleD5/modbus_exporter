package handler

import (
	"net/http"

	"github.com/NobleD5/modbus_exporter/pkg/structures"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/server"

	yaml "gopkg.in/yaml.v2"
)

// Healthy return OK when probed
func Healthy() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "OK", http.StatusOK)
	})
}

// Ready return OK when probed
func Ready() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "OK", http.StatusOK)
	})
}

// ReloadConfig reload config
func ReloadConfig(reload chan bool, logger log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		level.Debug(logger).Log("msg", "Requesting app's configuration reload!")
		http.Error(w, "Accepted", http.StatusAccepted)
		reload <- true
	})
}

// Quit send boolean chan to goroutine for gracefully quit application
func Quit(quit chan bool, logger log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		level.Debug(logger).Log("msg", "Requesting app's termination.")
		close(quit)
	})
}

// ShowConfig show config
func ShowConfig(safeConfig *structures.SafeConfig, logger log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		safeConfig.RLock()
		config, err := yaml.Marshal(safeConfig.C)
		safeConfig.RUnlock()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			level.Error(logger).Log("msg", "Error marshalling configuration", "err", err.Error())
			return
		}
		w.Write(config)
	})
}

// Static serves the static files from the provided http.FileSystem.
func Static(root http.FileSystem, prefix string) http.Handler {

	if prefix == "/" {
		prefix = ""
	}

	handler := server.StaticFileServer(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = r.URL.Path[len(prefix):]
		handler.ServeHTTP(w, r)
	})
}
