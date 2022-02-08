package handler

import (
	"net/http"
	"time"

	"github.com/NobleD5/modbus_exporter/pkg/collector"
	"github.com/NobleD5/modbus_exporter/pkg/structures"
	"github.com/NobleD5/modbus_exporter/pkg/workload"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Modbus serves the modbus page.
//
// The returned handler is already instrumented for Prometheus.
func Modbus(
	safeConfig *structures.SafeConfig,
	logger log.Logger,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var (
			// Metrics about the exporter itself
			modbusDuration = prometheus.NewSummaryVec(
				prometheus.SummaryOpts{
					Name: "modbus_collection_duration_seconds",
					Help: "Duration of collections by the Modbus exporter",
				},
				[]string{"config"},
			)
			modbusRequestErrors = prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: "modbus_request_errors_total",
					Help: "Errors in requests to the Modbus exporter",
				},
			)
		)

		// Checking target
		target := r.URL.Query().Get("target")
		if target == "" {
			level.Error(logger).Log("msg", "'target' parameter must be specified")
			http.Error(w, "'target' parameter must be specified", http.StatusBadRequest)
			modbusRequestErrors.Inc()
			return
		}

		// Checking config
		configName := r.URL.Query().Get("config")
		if configName == "" {
			level.Error(logger).Log("msg", "'config' parameter must be specified")
			http.Error(w, "'config' parameter must be specified", http.StatusBadRequest)
			modbusRequestErrors.Inc()
			return
		}

		//
		safeConfig.RLock()
		config, ok := (*(safeConfig.C))[configName]
		safeConfig.RUnlock()
		if !ok {
			level.Error(logger).Log("msg", "Unknown", "config", configName)
			http.Error(w, "unknown config", http.StatusBadRequest)
			modbusRequestErrors.Inc()
			return
		}

		level.Debug(logger).Log("msg", "Scrape target with config",
			"target", target,
			"config", configName,
		)

		workload, error := workload.PrepareConfig(config, logger)
		if error != nil {
			level.Error(logger).Log("msg", "Error preparing workload", "error", error.Error())
			http.Error(w, "error preparing workload", http.StatusBadRequest)
			return
		}

		start := time.Now()
		registry := prometheus.NewRegistry()
		modbusCollector := collector.NewModbusCollector(target, workload, logger)
		registry.MustRegister(modbusCollector)

		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
		duration := float64(time.Since(start).Seconds())
		modbusDuration.WithLabelValues(configName).Observe(duration)

		level.Debug(logger).Log("msg", "Scrape of target with config",
			"target", target,
			"config", configName,
			"msg", "took seconds",
			"duration", duration,
		)

	})

}
