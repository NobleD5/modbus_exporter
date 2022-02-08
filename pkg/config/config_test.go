package config

import (
	"errors"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func TestUnmarshallConf(t *testing.T) {

	var (
		AwaitingUnmarshallErr = errors.New(
			"error unmarshaling given yaml input: yaml: unmarshal errors:\n" +
				"  line 1: cannot unmarshal !!int `1` into structures.Params\n" +
				"  line 2: cannot unmarshal !!int `2` into structures.Params",
		)
	)

	logger := log.NewLogfmtLogger(os.Stdout)

	logger = level.NewFilter(logger, level.AllowError())
	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	// ---------------------------------------------------------------------------
	//  CASE:
	// ---------------------------------------------------------------------------
	_, err := UnmarshallConf("a: 1\nb: 2", logger)
	if err != nil && err.Error() == AwaitingUnmarshallErr.Error() {
		t.Logf("UnmarshallConf() : Test 1 PASSED, \nawating and got this: %s", AwaitingUnmarshallErr.Error())
	} else {
		t.Errorf("UnmarshallConf() : Test 1 FAILED, \nawaiting: \"%s\", \ngot: \"%s\"", AwaitingUnmarshallErr.Error(), err.Error())
	}

}

func TestLoadFile(t *testing.T) {

	var (
		// AwaitingUnmarshallErr = errors.New("error loading read file (content) into config: error unmarshaling given yaml input: yaml: unmarshal errors:\n          line 3: cannot unmarshal !!seq into structures.Params")
		AwaitingNoFile = errors.New("error reading given filename (): open : no such file or directory")
	)

	logger := log.NewLogfmtLogger(os.Stdout)

	logger = level.NewFilter(logger, level.AllowError())
	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	// ---------------------------------------------------------------------------
	//  CASE:
	// ---------------------------------------------------------------------------
	_, err := LoadFile("", logger)
	if err != nil && err.Error() == AwaitingNoFile.Error() {
		t.Logf("LoadFile() : Test 1 PASSED, \nawating and got this: \"%s\"", AwaitingNoFile.Error())
	} else {
		t.Errorf("LoadFile() : Test 1 FAILED, \nawating: \"%s\", \ngot: \"%s\"", AwaitingNoFile.Error(), err.Error())
	}

}

func TestLoadDirectory(t *testing.T) {

	var (
	// AwaitingUnmarshallErr = errors.New("error loading read file (content) into config: error unmarshaling given yaml input: yaml: unmarshal errors:\n          line 3: cannot unmarshal !!seq into structures.Params")
	// AwaitingNoFile = errors.New("error reading given directory: open ../testdata/testdir/noexist.yaml: no such file or directory")
	)

	logger := log.NewLogfmtLogger(os.Stdout)

	logger = level.NewFilter(logger, level.AllowError())
	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	// ---------------------------------------------------------------------------
	//  CASE: found dir'n'files and successfully loading them
	// ---------------------------------------------------------------------------
	_, err := LoadDirectory("../testdata/testdir", logger)
	if err != nil {
		t.Errorf("LoadDirectory() : Test 1 FAILED, %s", err)
	} else {
		t.Log("LoadDirectory() : Test 1 PASSED.")
	}

	// ---------------------------------------------------------------------------
	//  CASE: got empty dir
	// ---------------------------------------------------------------------------
	_, err = LoadDirectory("", logger)
	if err != nil {
		t.Log("LoadDirectory() : Test 2 PASSED")
	} else {
		t.Error("LoadDirectory() : Test 2 FAILED, awaiting error")
	}

}

func TestLoad(t *testing.T) {

	var (
		AwaitingNoDir    = errors.New("error while os.Stat: stat ../testdat: no such file or directory")
		AwaitingEmptyDir = errors.New("error while LoadDirectory: directory is empty")
		AwaitingWrYAML   = errors.New(
			"error loading read file (content) into config: error unmarshaling given yaml input: yaml: unmarshal errors:\n" +
				"  line 3: cannot unmarshal !!seq into structures.Params",
		)
	)

	logger := log.NewLogfmtLogger(os.Stdout)

	logger = level.NewFilter(logger, level.AllowError())
	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	// ---------------------------------------------------------------------------
	//  CASE:
	// ---------------------------------------------------------------------------
	_, err := Load("../testdata/testdir", logger)
	if err != nil {
		t.Errorf("Load() : Test 1 FAILED, %s", err)
	} else {
		t.Log("Load() : Test 1 PASSED")
	}

	// ---------------------------------------------------------------------------
	//  CASE:
	// ---------------------------------------------------------------------------
	_, err = Load("../testdat", logger)
	if err != nil && err.Error() == AwaitingNoDir.Error() {
		t.Logf("Load() : Test 2 PASSED, \nawating and got this: \"%s\"", AwaitingNoDir.Error())
	} else {
		t.Errorf("Load() : Test 2 FAILED, \nawating: \"%s\", \ngot: \"%s\"", AwaitingNoDir.Error(), err.Error())
	}

	// ---------------------------------------------------------------------------
	//  CASE:
	// ---------------------------------------------------------------------------
	_, err = Load("../testdata/emptydir", logger)
	if err != nil && err.Error() == AwaitingEmptyDir.Error() {
		t.Logf("Load() : Test 3 PASSED, \nawating and got this: \"%s\"", AwaitingEmptyDir.Error())
	} else {
		t.Errorf("Load() : Test 3 FAILED, \nawating: \"%s\", \ngot: \"%s\"", AwaitingNoDir.Error(), err.Error())
	}

	// ---------------------------------------------------------------------------
	//  CASE:
	// ---------------------------------------------------------------------------
	_, err = Load("../testdata/valid_modbus_conf.yaml", logger)
	if err != nil {
		t.Errorf("Load() : Test 4 FAILED, %s", err)
	} else {
		t.Log("Load() : Test 4 PASSED")
	}

	// ---------------------------------------------------------------------------
	//  CASE:
	// ---------------------------------------------------------------------------
	_, err = Load("../testdata/invalid_modbus_conf.yaml", logger)
	if err != nil && err.Error() == AwaitingWrYAML.Error() {
		t.Logf("Load() : Test 5 PASSED, \nawating and got this: \"%s\"", AwaitingWrYAML.Error())
	} else {
		t.Errorf("Load() : Test 5 FAILED, \nawating: \"%s\", \ngot: \"%s\"", AwaitingWrYAML.Error(), err.Error())
	}

	// Test 6 --------------------------------------------------------------------
	// labels:
	//   prompt": "dfdf"
	// _, err = Load("../testdata/invalid_modbus_conf.yaml", logger)
	// if err != nil && err.Error() == AwaitingWrYAML.Error() {
	// 	t.Logf("Load() : Test 5 PASSED, \nawating and got this: \"%s\"", AwaitingWrYAML.Error())
	// } else {
	// 	t.Errorf("Load() : Test 5 FAILED, \nawating: \"%s\", \ngot: \"%s\"", AwaitingWrYAML.Error(), err.Error())
	// }

}
