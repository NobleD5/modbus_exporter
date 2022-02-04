package main

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"gitlab.dc.miran.ru/nuzhin/modbus_exporter/pkg/config"
	"gitlab.dc.miran.ru/nuzhin/modbus_exporter/pkg/handler"
	"gitlab.dc.miran.ru/nuzhin/modbus_exporter/pkg/logger"
	"gitlab.dc.miran.ru/nuzhin/modbus_exporter/pkg/structures"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/route"
	"github.com/prometheus/common/version"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// Types declaration section ---------------------------------------------------

// Constants declaration section -----------------------------------------------
const (
	showError = string("ERROR")
	showWarn  = string("WARN")
	showInfo  = string("INFO")
	showDebug = string("DEBUG")
)

// Variables declaration section -----------------------------------------------
var (

	// KINGPIN flags
	app = kingpin.New(filepath.Base(os.Args[0]), "The Modbus Exporter")

	// main config flag
	configFile = app.Flag("config.file", "Config file (or directory with multiple files) which contains list of devices with modbus-registers.").Required().String()
	// app options flags
	listenAddress      = app.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9700").String()
	metricsPath        = app.Flag("web.metrics-path", "Path under which to expose MODBUS metrics.").Default("/modbus").String()
	telemetryPath      = app.Flag("web.telemetry-path", "Path under which to expose telemetry metrics.").Default("/metrics").String()
	externalURL        = app.Flag("web.external-url", "The URL under which the Pushgateway is externally reachable.").Default("").URL()
	routePrefix        = app.Flag("web.route-prefix", "Prefix for the internal routes of web endpoints. Defaults to the path of --web.external-url.").Default("").String()
	insecureSkipVerify = app.Flag("web.insecure", "Skip verification in requests.").Bool()
	logLevel           = app.Flag("log.level", "Log level. Valid are next values: DEBUG, INFO, WARN, ERROR.").Default("INFO").String()

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

	assets http.FileSystem = http.Dir("../resources")
)

func init() {

	prometheus.MustRegister(modbusDuration)
	prometheus.MustRegister(modbusRequestErrors)
	prometheus.Register(version.NewCollector("modbus_exporter"))

}

func main() {

	app.Version(version.Print("modbus_exporter"))
	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))

	logger := logger.SetupLogger(*logLevel)

	*routePrefix = computeRoutePrefix(*routePrefix, *externalURL)
	externalPathPrefix := computeRoutePrefix("", *externalURL)

	level.Info(logger).Log(
		"msg", "KINGPIN would use this provided core flags:",
		"config", *configFile,
		"metrics", *metricsPath,
		"telemetry", *telemetryPath,
		"url", *externalURL,
	)

	level.Info(logger).Log(
		"msg", "KINGPIN would use this provided secondary flags:",
		"prefix", externalPathPrefix,
		"route", *routePrefix,
		"listen", *listenAddress,
		"logs", *logLevel,
	)

	level.Info(logger).Log("msg", "Starting MODBUS-exporter", "version", version.Info())
	level.Info(logger).Log("build_context", version.BuildContext())

	// flags is used to show command line flags on the status page.
	// Kingpin default flags are excluded as they would be confusing.
	flags := map[string]string{}
	boilerplateFlags := kingpin.New("", "").Version("")
	for _, f := range app.Model().Flags {
		if boilerplateFlags.GetFlag(f.Name) == nil {
			flags[f.Name] = f.Value.String()
		}
	}

	var err error
	safeConfig := structures.NewSafeConfig(&structures.Config{})

	safeConfig.C, err = config.Load(*configFile, logger)
	if err != nil {
		level.Error(logger).Log("msg", "Error while loading config file (*.yaml)", "err", err.Error())
		os.Exit(1)
	}

	// for config := range *safeConfig.C {
	// 	modbusDuration.WithLabelValues(config)
	// }

	// hup := make(chan os.Signal)
	// reload = make(chan chan error)

	reload := make(chan bool)
	quit := make(chan bool)

	// signal.Notify(hup, syscall.SIGHUP)

	go reloadConfigOnReload(safeConfig, reload, logger)

	// Creating new route
	r := route.New()

	// -------------------------- MISC endpoints  --------------------------------
	r.Get((*routePrefix + "/-/healthy"), handler.Healthy().ServeHTTP)                   // Healthy endpoint
	r.Get((*routePrefix + "/-/ready"), handler.Ready().ServeHTTP)                       // Ready endpoint
	r.Get((*routePrefix + "/config"), handler.ShowConfig(safeConfig, logger).ServeHTTP) // Show configuration endpoint

	r.Get((*routePrefix + "/-/reload"), func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only POST or PUT requests allowed."))
	})
	r.Post((*routePrefix + "/-/reload"), handler.ReloadConfig(reload, logger).ServeHTTP) // Endpoint to reload configuration
	r.Put((*routePrefix + "/-/reload"), handler.ReloadConfig(reload, logger).ServeHTTP)

	r.Get((*routePrefix + "/-/quit"), func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only POST or PUT requests allowed."))
	})
	r.Post((*routePrefix + "/-/quit"), handler.Quit(quit, logger).ServeHTTP) // Endpoint for application graceful quit
	r.Put((*routePrefix + "/-/quit"), handler.Quit(quit, logger).ServeHTTP)
	// ---------------------------------------------------------------------------

	// --------- MAIN endpoint for MODBUS metrics scrapes ------------------------
	r.Get((*routePrefix + *metricsPath), handler.Modbus(safeConfig, logger).ServeHTTP)
	// ---------------------------------------------------------------------------

	// -------- MAIN endpoint for telemetry metrics scrapes ----------------------
	r.Get((*routePrefix + *telemetryPath), promhttp.Handler().ServeHTTP)
	// ---------------------------------------------------------------------------

	// -------- MAIN endpoint for status screen for modbus exporter itself -------
	r.Get((*routePrefix + "/static/*filepath"), handler.Static(assets, *routePrefix).ServeHTTP)
	r.Get((*routePrefix + "/"), handler.Status(assets, flags, externalPathPrefix, logger).ServeHTTP)
	// ---------------------------------------------------------------------------

	level.Info(logger).Log("listen_address", *listenAddress)
	listener, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/", r)

	go closeListenerOnQuit(listener, quit, logger)

	err = (&http.Server{Addr: *listenAddress, Handler: mux}).Serve(listener)
	level.Error(logger).Log("msg", "HTTP server stopped", "err", err.Error())
}

// reloadConfig re-apply safe config
func reloadConfig(sc *structures.SafeConfig, newConf string, logger log.Logger) (error error) {

	config, error := config.Load(newConf, logger)
	if error != nil {
		level.Error(logger).Log("err", error.Error())
		return error
	}

	sc.Lock()
	sc.C = config
	sc.Unlock()

	level.Info(logger).Log("msg", "Loaded configuration file")
	return nil
}

// computeRoutePrefix returns the effective route prefix based on the
// provided flag values for --web.route-prefix and
// --web.external-url. With prefix empty, the path of externalURL is
// used instead. A prefix "/" results in an empty returned prefix. Any
// non-empty prefix is normalized to start, but not to end, with "/".
func computeRoutePrefix(prefix string, externalURL *url.URL) string {

	if prefix == "" {
		prefix = externalURL.Path
	}

	if prefix == "/" {
		prefix = ""
	}

	if prefix != "" {
		prefix = "/" + strings.Trim(prefix, "/")
	}

	return prefix
}

// closeListenerOnQuite closes the provided listener upon closing the provided
// 'quit' or upon receiving a SIGINT or SIGTERM.
func closeListenerOnQuit(listener net.Listener, quit <-chan bool, logger log.Logger) {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-signals:
		level.Warn(logger).Log("msg", "Received SIGINT/SIGTERM; exiting gracefully...")
		break
	case <-quit:
		level.Warn(logger).Log("msg", "Received termination request via web service, exiting gracefully...")
		break
	}

	listener.Close()
}

// reloadConfigOnReload will try to reload current config with a new one upon receiving 'reload'
// or upon receiving a SIGHUP.
func reloadConfigOnReload(safeConfig *structures.SafeConfig, reload <-chan bool, logger log.Logger) {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)

	for {
		select {
		case <-signals:
			level.Warn(logger).Log("msg", "Received SIGHUP, trying to reload configuration...")
			if err := reloadConfig(safeConfig, *configFile, logger); err != nil {
				level.Error(logger).Log("msg", "Error reloading configuration", "err", err.Error())
			}
		case <-reload:
			level.Warn(logger).Log("msg", "Received reload request via web service, trying to reload configuration...")
			if err := reloadConfig(safeConfig, *configFile, logger); err != nil {
				level.Error(logger).Log("msg", "Error reloading configuration", "err", err.Error())
			}
		}

	}

}
