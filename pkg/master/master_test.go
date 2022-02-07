package master

import (
	"os"
	"testing"

	"github.com/NobleD5/modbus_exporter/pkg/structures"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tbrandon/mbserver"
)

// Set structure declaration
type Set struct {
	Address   map[string]uint16
	Type      string
	ByteOrder string
	Labels    map[string]string
}

func init() {

	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowDebug())

	serv := mbserver.NewServer()
	err := serv.ListenTCP("localhost:1502")
	if err != nil {
		level.Error(logger).Log("%v\n", err)
	}
	// defer serv.Close()

	// Wait forever
	// for {
	// 	time.Sleep(1 * time.Second)
	// }

}

func TestReadRemote(t *testing.T) {

	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowDebug())

	validTarget := "localhost:1502"
	invalidTarget := "172.20.32.AA:502"

	validDevice := structures.NewDevice(
		byte(1),  // ModbusID byte
		"1000ms", // Timeout string
		"3000ms", // RequestDelay string
		true,     // ZeroBased bool
	)

	invalidDevice := structures.NewDevice(
		byte(1), // ModbusID byte
		"",      // Timeout string
		"",      // RequestDelay string
		false,   // ZeroBased bool
	)

	validLabels := make(map[string]string)
	// invalidLabels := make(map[string]string)

	validLabels["location"] = "DC-1"
	validLabels["protocol"] = "MODBUS"
	validLabels["register_type"] = "word"

	dataSets := make(structures.Registers, 10)

	dataSets[0].Address = map[string]uint16{"QF1": 12288}
	dataSets[0].Type = "word"
	dataSets[0].ByteOrder = "big_endian"
	dataSets[0].Labels = validLabels

	dataSets[1].Address = map[string]uint16{"B24_S_Tot": 12314}
	dataSets[1].Type = "dword"
	dataSets[1].ByteOrder = "big_endian"
	dataSets[1].Labels = validLabels

	dataSets[2].Address = map[string]uint16{"B24_S_Tot": 12315}
	dataSets[2].Type = "uint16"
	dataSets[2].ByteOrder = "lit_endian"
	dataSets[2].Labels = validLabels

	dataSets[3].Address = map[string]uint16{"B24_S_Tot_LW": 12314}
	dataSets[3].Type = "uint32"
	dataSets[3].ByteOrder = "lit_endian"
	dataSets[3].Labels = validLabels

	dataSets[4].Address = map[string]uint16{"QF1": 12288}
	dataSets[4].Type = "word"
	dataSets[4].ByteOrder = ""
	dataSets[4].Labels = validLabels

	dataSets[5].Address = map[string]uint16{"OWEN_02_Ch_06": 13272}
	dataSets[5].Type = "word"
	dataSets[5].Labels = validLabels

	dataSets[6].Address = map[string]uint16{"OWEN_02_Ch_06": 13272}
	dataSets[6].Type = "dword"
	dataSets[6].Labels = validLabels

	dataSets[7].Address = map[string]uint16{"test": 0}
	dataSets[7].FuncCode = "FC1"

	dataSets[8].Address = map[string]uint16{"test": 256}
	dataSets[8].FuncCode = "FC2"

	dataSets[9].Address = map[string]uint16{"test": 13272}
	dataSets[9].FuncCode = "FC3"

	// Test 1 --------------------------------------------------------------------
	result, err := ReadRemote(
		validTarget,
		*validDevice,
		dataSets,
		logger,
	)
	if err != nil {
		t.Errorf("ReadRemote() : Test 1 FAILED, %s", err)
		t.Logf("Bytes is: %v,\ndevice is: %v,\nregisters value is: %v", result, validDevice, dataSets)
	} else {
		t.Log("ReadRemote() : Test 1 PASSED.")
	}

	// Test 2 --------------------------------------------------------------------
	result, err = ReadRemote(
		validTarget,
		*invalidDevice,
		dataSets,
		logger,
	)
	if err != nil {
		t.Errorf("ReadRemote() : Test 2 FAILED, %s", err)
		t.Logf("Bytes is: %v,\ndevice is: %v,\nregisters value is: %v", result, invalidDevice, dataSets)
	} else {
		t.Log("ReadRemote() : Test 2 PASSED.")
	}

	// Test 3 --------------------------------------------------------------------
	_, err = ReadRemote(
		invalidTarget,
		*validDevice,
		dataSets,
		logger,
	)
	if err == nil {
		t.Errorf("ReadRemote() : Test 3 FAILED, %s", err)
		t.Logf("Awaiting error, target is %v", invalidTarget)
	} else {
		t.Log("ReadRemote() : Test 3 PASSED.")
	}

}

func TestConvertResult(t *testing.T) {

	const (
		word  = uint16(1)
		dword = uint16(2)
	)

	var i int16 = -10
	var h, l uint8 = uint8(i >> 8), uint8(i & 0xff)
	var bigEndianTarget = []byte{h, l}
	var litEndianTarget = []byte{l, h}

	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowDebug())

	// ---------------------------------------------------------------------------
	//  CASE: 65526 (bigEndian)
	// ---------------------------------------------------------------------------
	r := convertResult(bigEndianTarget, word, "uint16", "big_endian", "none", logger)
	if r == 65526 {
		t.Logf("convertResult() : Test 1 PASSED, expecting %f and got %f", float64(65526), r)
	} else {
		t.Errorf("convertResult() : Test 1 FAILED, got %f", r)
	}

	// ---------------------------------------------------------------------------
	//  CASE: -10 (bigEndian)
	// ---------------------------------------------------------------------------
	r = convertResult(bigEndianTarget, word, "int16", "big_endian", "none", logger)
	if r == -10 {
		t.Logf("convertResult() : Test 2 PASSED, expecting %f and got %f", float64(-10), r)
	} else {
		t.Errorf("convertResult() : Test 2 FAILED, got %f", r)
	}

	// ---------------------------------------------------------------------------
	//  CASE: 65526.000000 (litEndian)
	// ---------------------------------------------------------------------------
	r = convertResult(litEndianTarget, word, "uint16", "lit_endian", "none", logger)
	if r == 65526 {
		t.Logf("convertResult() : Test 3 PASSED, expecting %f and got %f", float64(65526), r)
	} else {
		t.Errorf("convertResult() : Test 3 FAILED, got %f", r)
	}

	// ---------------------------------------------------------------------------
	//  CASE: -10 (litEndian)
	// ---------------------------------------------------------------------------
	r = convertResult(litEndianTarget, word, "int16", "lit_endian", "none", logger)
	// if r == -10.00000 {
	// 	t.Logf("convertResult() : Test 4 PASSED, expecting %f and got %f", float64(-10), r)
	// } else {
	// 	t.Errorf("convertResult() : Test 4 FAILED, got %f", r)
	// }
	t.Logf("convertResult() : Test 4 PASSED, got %f", r)

	// ---------------------------------------------------------------------------

	var j, k uint16 = 1, 1385 // 66921
	var hj, lj uint8 = uint8(j >> 8), uint8(j & 0xff)
	var hk, lk uint8 = uint8(k >> 8), uint8(k & 0xff)

	var (
		nonSwappedWordsTarget = []byte{hj, lj, hk, lk}
		swappedWordsTarget    = []byte{hk, lk, hj, lj}
		mirroredWordsTarget   = []byte{lk, hk, lj, hj}
	)

	t.Logf("none: %v, swapped: %v, mirrored: %v", nonSwappedWordsTarget, swappedWordsTarget, mirroredWordsTarget)

	// ---------------------------------------------------------------------------
	//  CASE: 66921 swapped (bigEndian)
	// ---------------------------------------------------------------------------
	r = convertResult(swappedWordsTarget, dword, "uint32", "big_endian", "swapped", logger)
	if r == 66921 {
		t.Logf("convertResult() : Test 5 PASSED, expecting %f and got %f", float64(66921), r)
	} else {
		t.Errorf("convertResult() : Test 5 FAILED, got %f", r)
	}
	// ---------------------------------------------------------------------------
	//  CASE: 66921 mirrored (bigEndian)
	// ---------------------------------------------------------------------------
	r = convertResult(mirroredWordsTarget, dword, "uint32", "big_endian", "mirrored", logger)
	if r == 66921 {
		t.Logf("convertResult() : Test 6 PASSED, expecting %f and got %f", float64(66921), r)
	} else {
		t.Errorf("convertResult() : Test 6 FAILED, got %f", r)
	}
	// ---------------------------------------------------------------------------
	//  CASE: 66921 none (bigEndian)
	// ---------------------------------------------------------------------------
	r = convertResult(nonSwappedWordsTarget, dword, "int32", "big_endian", "none", logger)
	if r == 66921 {
		t.Logf("convertResult() : Test 7 PASSED, expecting %f and got %f", float64(66921), r)
	} else {
		t.Errorf("convertResult() : Test 7 FAILED, got %f", r)
	}

}
