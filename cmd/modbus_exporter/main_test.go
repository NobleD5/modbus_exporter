package main

import (
	"fmt"
	"net"
	// "net/http"
	// "net/http/httptest"
	"net/url"
	"os"
	// "os/signal"
	// "syscall"
	"testing"

	"gitlab.dc.miran.ru/nuzhin/modbus_exporter/pkg/config"
	// "gitlab.dc.miran.ru/nuzhin/modbus_exporter/pkg/handler"
	"gitlab.dc.miran.ru/nuzhin/modbus_exporter/pkg/logger"
	"gitlab.dc.miran.ru/nuzhin/modbus_exporter/pkg/structures"

	// "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func TestMainFunc(t *testing.T) {

	quit := make(chan bool)

	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				os.Args = []string{"",
					"--config.file=valid_modbus_conf.yaml",
					"--log.level=DEBUG"}
				main()
			}
		}
	}()

	quit <- true

}

func TestReloadConfig(t *testing.T) {

	logger := logger.SetupLogger("DEBUG")

	sc := structures.NewSafeConfig(&structures.Config{})

	var err error
	sc.C, err = config.Load("valid_modbus_conf.yaml", logger)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	// Test 1 --------------------------------------------------------------------
	err = reloadConfig(sc, "valid_modbus_conf.yaml", logger)
	if err != nil {
		t.Errorf("reloadConfig() : Test 1 FAILED, %s", err)
	} else {
		t.Log("reloadConfig() : Test 1 PASSED.")
	}
}

func TestComputeRoutePrefix(t *testing.T) {

	prefixSlice := []string{
		"",
		"/",
		"prefix",
	}

	url, _ := url.Parse("example.com")

	for _, prefix := range prefixSlice {
		computePrefix := computeRoutePrefix(prefix, url)
		fmt.Println(computePrefix)
	}

}

func TestCloseListenerOnQuit(t *testing.T) {

	logger := logger.SetupLogger("DEBUG")
	quit := make(chan bool)

	listener, err := net.Listen("tcp", ":2000")
	if err != nil {
		level.Error(logger).Log("err", err)
	}

	defer listener.Close()

	// Test 1 --------------------------------------------------------------------
	go func() {
		closeListenerOnQuit(listener, quit, logger)
	}()

	quit <- true
	t.Log("closeListenerOnQuit() : Test 1 PASSED.")

}

func TestReloadConfigOnReload(t *testing.T) {

	logger := logger.SetupLogger("DEBUG")
	reload := make(chan bool)

	sc := structures.NewSafeConfig(&structures.Config{})

	go func() {
		reloadConfigOnReload(sc, reload, logger)
	}()

	// Test 1 --------------------------------------------------------------------
	reload <- true
	t.Log("reloadConfigOnReload() : Test 1 PASSED.")
}
