package collector

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/NobleD5/modbus_exporter/pkg/structures"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/tbrandon/mbserver"
)

// Set structure declaration
type Set struct {
	MetricType   string
	MetricSiName string
	Value        float64
	Labels       map[string]string
}

const (
	validTarget   = "localhost:1902"
	invalidTarget = "loc.AA"
)

func init() {

	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowDebug())

	serv := mbserver.NewServer()
	err := serv.ListenTCP(validTarget)
	if err != nil {
		level.Error(logger).Log("%v\n", err)
	}

}

func TestRegisterToSamples(t *testing.T) {

	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowDebug())

	var (
		validLabels   map[string]string
		invalidLabels map[string]string
	)

	validLabels = make(map[string]string)
	invalidLabels = make(map[string]string)

	validLabels["location"] = "DC-2"
	validLabels["protocol"] = "MODBUS"

	invalidLabels[""] = ""

	dataSets := make([]Set, 5)

	dataSets[0].MetricType = "uint16"
	dataSets[0].MetricSiName = "current"
	dataSets[0].Value = float64(10)
	dataSets[0].Labels = validLabels

	dataSets[1].MetricType = "word"
	dataSets[1].MetricSiName = ""
	dataSets[1].Value = float64(5)
	dataSets[1].Labels = invalidLabels

	dataSets[2].MetricType = "int32"
	dataSets[2].MetricSiName = "energy"
	dataSets[2].Value = float64(4.206362624e+09)
	dataSets[2].Labels = validLabels

	dataSets[3].MetricType = "counter"

	dataSets[4].MetricType = "int32"

	// Do in cycle all data sets tests
	for _, set := range dataSets {

		set := set

		samples, err := RegisterToSamples(
			set.MetricType,
			set.MetricSiName,
			set.Value,
			set.Labels,
			logger,
		)
		if err != nil {
			t.Errorf("RegisterToSamples() : test FAILED, %s", err)
			t.Logf("Sample is: %v, labels set is: %v", samples, set.Labels)
		} else {
			t.Log("RegisterToSamples() : test PASSED.")
		}
	}

}

func TestScrapeTarget(t *testing.T) {

	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowDebug())

	validDevice := structures.NewDevice(1, "1000ms", "3000ms", true)

	validDataUnitA := structures.NewDataUnit(
		map[string]uint16{"dec": uint16(331)},
		float64(0),
		"QF1", "bool", "word", "big_endian", "",
		map[string]string{"protocol": "MODBUS", "register_type": "word"},
	)

	validDataUnitB := structures.NewDataUnit(
		map[string]uint16{"dec": uint16(351)},
		float64(0),
		"QF2", "bool", "int32", "big_endian", "",
		map[string]string{"protocol": "MODBUS", "register_type": "word"},
	)

	validWorkload := map[structures.Device]structures.Registers{
		*validDevice: structures.Registers{*validDataUnitA, *validDataUnitB},
	}

	// ---------------------------------------------------------------------------
	//  CASE: Scrape valid target
	// ---------------------------------------------------------------------------
	data, err := ScrapeTarget(validTarget, validWorkload, logger)
	if err != nil {
		t.Errorf("ScrapeTarget() : Test 1 FAILED, %s", err)
		t.Logf("Data unit is: %v\n", data)
	} else {
		t.Log("ScrapeTarget() : Test 1 PASSED.")
	}

	// ---------------------------------------------------------------------------
	//  CASE: Scrape invalid target
	// ---------------------------------------------------------------------------
	_, err = ScrapeTarget(invalidTarget, validWorkload, logger)
	if err == nil {
		t.Errorf("ScrapeTarget() : Test 2 FAILED, %s", err)
	} else {
		t.Log("ScrapeTarget() : Test 2 PASSED.")
	}

}

func TestCollector(t *testing.T) {

	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowDebug())

	validDevice := structures.NewDevice(1, "1000ms", "3000ms", true)

	validDataUnitA := structures.NewDataUnit(
		map[string]uint16{"dec": uint16(257)},
		float64(0),
		"QF1", "bool", "word", "big_endian", "",
		map[string]string{"vendor": "WAGO", "device": "PLC", "protocol": "MODBUS", "register_type": "word"},
	)

	validDataUnitB := structures.NewDataUnit(
		map[string]uint16{"dec": uint16(256)},
		float64(0),
		"TR01_T", "temperature", "int16", "", "",
		map[string]string{"vendor": "WAGO", "device": "PLC", "protocol": "MODBUS", "register_type": "int16"},
	)

	validDataUnitC := structures.NewDataUnit(
		map[string]uint16{"dec": uint16(250)},
		float64(0),
		"TR02_T", "temperature", "int16", "", "",
		map[string]string{"vendor": "WAGO", "device": "PLC", "register_type": "int16", "prompt": "Температура ангара"},
	)

	validDataUnitD := structures.NewDataUnit(
		map[string]uint16{"dec": uint16(12000)},
		float64(0),
		"TR03_T", "temperature", "int16", "", "",
		map[string]string{"register_type": "int16"},
	)

	validWorkloadA := map[structures.Device]structures.Registers{
		*validDevice: structures.Registers{*validDataUnitA, *validDataUnitB, *validDataUnitC},
	}
	validWorkloadB := map[structures.Device]structures.Registers{
		*validDevice: structures.Registers{*validDataUnitD},
	}

	registry := prometheus.NewRegistry()

	collectorA := NewModbusCollector(
		validTarget,
		validWorkloadA,
		logger,
	)

	collectorB := NewModbusCollector(
		invalidTarget,
		validWorkloadB,
		logger,
	)

	registry.MustRegister(collectorA)
	registry.MustRegister(collectorB)

	s := strings.NewReader(
		"# HELP modbus_bool metric.Help\n# TYPE modbus_bool gauge\nmodbus_bool{device=\"PLC\",protocol=\"MODBUS\",register_name=\"QF1\",register_type=\"word\",vendor=\"WAGO\"} 0\n\n" +
			"# HELP modbus_temperature metric.Help\n# TYPE modbus_temperature gauge\nmodbus_temperature{device=\"PLC\",protocol=\"MODBUS\",register_name=\"TR01_T\",register_type=\"int16\",vendor=\"WAGO\"} 0\n\n" +
			"modbus_temperature{device=\"PLC\",prompt=\"Температура ангара\",register_name=\"TR02_T\",register_type=\"int16\",vendor=\"WAGO\"} 0\n\n",
	)
	r := io.LimitReader(s, 1000)

	// ---------------------------------------------------------------------------
	//  CASE: CollectAndCompare
	// ---------------------------------------------------------------------------
	// err := testutil.CollectAndCompare(collectorA, r, "modbus_bool", "modbus_temperature")
	// if err != nil {
	// 	t.Errorf("CollectAndCompare() : Test 1 FAILED, %s", err.Error())
	// } else {
	// 	t.Log("CollectAndCompare() : Test 1 PASSED.")
	// }

	// ---------------------------------------------------------------------------
	//  CASE: CollectAndCompare
	// ---------------------------------------------------------------------------
	err := testutil.CollectAndCompare(collectorB, r)
	if err != nil {
		t.Log("CollectAndCompare() : Test 2 PASSED.")
	} else {
		t.Errorf("CollectAndCompare() : Test 2 FAILED, %s", err.Error())
	}

}
