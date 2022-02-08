package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"

	"github.com/NobleD5/modbus_exporter/pkg/config"
	"github.com/NobleD5/modbus_exporter/pkg/logger"
	"github.com/NobleD5/modbus_exporter/pkg/structures"

	// "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func TestMisc(t *testing.T) {

	logger := logger.SetupLogger("DEBUG")
	safeConfig := structures.NewSafeConfig(&structures.Config{})

	reload := make(chan bool)
	quit := make(chan bool)

	var err error
	var assets http.FileSystem = http.Dir("../resources")

	safeConfig.C, err = config.Load("../testdata/valid_modbus_conf.yaml", logger)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGHUP)
		for {
			select {
			case <-signals:
				level.Warn(logger).Log("msg", "Received SIGHUP, trying to reload configuration...")
			case <-reload:
				level.Warn(logger).Log("msg", "Received reload request via web service, trying to reload configuration...")
			}
		}
	}()

	go func() {
		select {
		case value, ok := <-reload:
			if ok {
				level.Info(logger).Log("msg", "Received channel", "value", value)
			} else {
				level.Info(logger).Log("msg", "Channel closed!")
			}
		default:
			level.Info(logger).Log("msg", "No value ready, moving on.")
		}
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("/-/ready", Ready().ServeHTTP)
	mux.HandleFunc("/-/healthy", Healthy().ServeHTTP)
	mux.HandleFunc("/-/reload", ReloadConfig(reload, logger).ServeHTTP)
	mux.HandleFunc("/-/quit", Quit(quit, logger).ServeHTTP)

	mux.HandleFunc("/config", ShowConfig(safeConfig, logger).ServeHTTP)

	mux.HandleFunc("/static", Static(assets, "/").ServeHTTP)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	configRoute, _ := url.Parse(ts.URL + "/config")
	staticRoute, _ := url.Parse(ts.URL + "/static")

	readyRoute, _ := url.Parse(ts.URL + "/-/ready")
	healthyRoute, _ := url.Parse(ts.URL + "/-/healthy")
	reloadRoute, _ := url.Parse(ts.URL + "/-/reload")
	quitRoute, _ := url.Parse(ts.URL + "/-/quit")

	// ---------------------------------------------------------------------------
	//  CASE: Getting route endpoint "/-/ready" without error and with proper response code
	// ---------------------------------------------------------------------------
	resp, err := http.Get(readyRoute.String())
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Errorf("Ready() : Test 1 FAILED, error: %v, expected status code: 200, got: %d", err, resp.StatusCode)
	} else {
		t.Logf("Ready() : Test 1 PASSED, expected %d and got status code: %d", http.StatusOK, resp.StatusCode)
	}

	// ---------------------------------------------------------------------------
	//  CASE: Getting route endpoint "/-/healthy" without error and with proper response code
	// ---------------------------------------------------------------------------
	resp, err = http.Get(healthyRoute.String())
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Errorf("Healthy() : Test 2 FAILED, error: %v, expected status code: 200, got: %d", err, resp.StatusCode)
	} else {
		t.Logf("Healthy() : Test 2 PASSED, expected %d and got status code: %d", http.StatusOK, resp.StatusCode)
	}

	// Test 3 --------------------------------------------------------------------
	// _, err = http.Get(reloadRoute.String())
	// if err != nil {
	// 	t.Errorf("ReloadConfig() : Test 3 FAILED, %s", err)
	// } else {
	// 	t.Log("ReloadConfig() : Test 3 PASSED.")
	// }

	// ---------------------------------------------------------------------------
	//  CASE: Successful POSTing "/-/reload" without error and with proper response code
	// ---------------------------------------------------------------------------
	resp, err = http.Post(reloadRoute.String(), "text/plain", strings.NewReader("body"))
	if err != nil || resp.StatusCode != http.StatusAccepted {
		t.Errorf("ReloadConfig() : Test 4 FAILED, error: %v, expected status code: 202, got: %d", err, resp.StatusCode)
	} else {
		t.Logf("ReloadConfig() : Test 4 PASSED, expected %d and got status code: %d", http.StatusAccepted, resp.StatusCode)
	}

	// ---------------------------------------------------------------------------
	//  CASE: Getting route endpoint "/config" without error
	// ---------------------------------------------------------------------------
	_, err = http.Get(configRoute.String())
	if err != nil {
		t.Errorf("ShowConfig() : Test 5 FAILED, %s", err)
	} else {
		t.Log("ShowConfig() : Test 5 PASSED.")
	}

	// ---------------------------------------------------------------------------
	//  CASE: Getting route endpoint "/static" without error
	// ---------------------------------------------------------------------------
	_, err = http.Get(staticRoute.String())
	if err != nil {
		t.Errorf("Static() : Test 6 FAILED, %s", err)
	} else {
		t.Log("Static() : Test 6 PASSED.")
	}

	// ---------------------------------------------------------------------------
	//  CASE: Successful POSTing "/-/quit" without error and with proper response code
	// ---------------------------------------------------------------------------
	resp, err = http.Post(quitRoute.String(), "text/plain", strings.NewReader("body"))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Errorf("Quit() : Test 7 FAILED, error: %v, expected status code: 200, got: %d", err, resp.StatusCode)
	} else {
		t.Logf("Quit() : Test 7 PASSED, expected %d and got status code: %d", http.StatusOK, resp.StatusCode)
	}

}
