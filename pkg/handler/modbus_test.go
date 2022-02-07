package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/NobleD5/modbus_exporter/pkg/config"
	"github.com/NobleD5/modbus_exporter/pkg/logger"
	"github.com/NobleD5/modbus_exporter/pkg/structures"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tbrandon/mbserver"
)

func init() {

	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowDebug())

	serv := mbserver.NewServer()
	err := serv.ListenTCP("localhost:2102")
	if err != nil {
		level.Error(logger).Log("%v\n", err)
	}

}

func TestModbus(t *testing.T) {

	logger := logger.SetupLogger("DEBUG")

	safeConfig := structures.NewSafeConfig(&structures.Config{})

	var err error
	safeConfig.C, err = config.Load("../testdata/valid_modbus_conf.yaml", logger)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/modbus", Modbus(safeConfig, logger).ServeHTTP)

	// mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
	// 	// res.Header().Set("Content-Type", "application/json")
	// 	// res.Header().Set("Location", "/REST/2.0/ticket/999999")
	// 	res.WriteHeader(http.StatusOK)
	// })

	ts := httptest.NewServer(mux)
	defer ts.Close()

	modbusValidRoute, _ := url.Parse(ts.URL + "/modbus" + "?config=PLC001&target=localhost%3A2102")
	modbusNoTargRoute, _ := url.Parse(ts.URL + "/modbus")
	modbusNoConfRoute, _ := url.Parse(ts.URL + "/modbus?target=localhost%3A2102")
	modbusUnkConfRoute, _ := url.Parse(ts.URL + "/modbus?config=PLC&target=localhost%3A2102")

	// Test 1 --------------------------------------------------------------------
	resp, err := http.Get(modbusValidRoute.String())
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Errorf("Modbus() : Test 1 FAILED, error: %v, expected status code: 200, got: %d", err, resp.StatusCode)
	} else {
		t.Logf("Modbus() : Test 1 PASSED, expected %d and got status code: %d", http.StatusOK, resp.StatusCode)
	}

	// Test 2 --------------------------------------------------------------------
	resp, err = http.Get(modbusNoTargRoute.String())
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Modbus() : Test 2 FAILED, error: %v, expected status code: 400, got: %d", err, resp.StatusCode)
	} else {
		t.Logf("Modbus() : Test 2 PASSED, expected %d and got status code: %d", http.StatusBadRequest, resp.StatusCode)
	}

	// Test 3 --------------------------------------------------------------------
	resp, err = http.Get(modbusNoConfRoute.String())
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Modbus() : Test 3 FAILED, error: %v, expected status code: 400, got: %d", err, resp.StatusCode)
	} else {
		t.Logf("Modbus() : Test 3 PASSED, expected %d and got status code: %d", http.StatusBadRequest, resp.StatusCode)
	}

	// Test 4 --------------------------------------------------------------------
	resp, err = http.Get(modbusUnkConfRoute.String())
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Modbus() : Test 4 FAILED, error: %v, expected status code: 400, got: %d", err, resp.StatusCode)
	} else {
		t.Logf("Modbus() : Test 4 PASSED, expected %d and got status code: %d", http.StatusBadRequest, resp.StatusCode)
	}

}
