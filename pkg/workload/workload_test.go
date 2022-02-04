package workload

import (
	"errors"
	"os"
	"testing"

	"gitlab.dc.miran.ru/nuzhin/modbus_exporter/pkg/structures"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func TestPrepareConfig(t *testing.T) {

	var (
		validConfig = &structures.Params{}
		errAwaiting = errors.New("strconv.ParseUint: parsing \"15F\": invalid syntax")
	)

	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowDebug())

	validConfig.DeviceTimeout = "1000ms"
	validConfig.DeviceModbusID = 1
	validConfig.DeviceRequestDelay = "3000ms"
	validConfig.DeviceLabels = map[string]string{"vendor": "VENDOR", "device": "PLC"}

	validRegisterA := structures.NewRegister("QF1", "bool", "word", "big_endian", "none",
		"dec#331",
		"",
		map[string]string{"protocol": "MODBUS", "register_type": "word"},
	)

	validRegisterB := structures.NewRegister("QF1", "bool", "wrong_value", "big_endian", "none",
		"hex#15F",
		"",
		map[string]string{"protocol": "MODBUS"},
	)

	invalidRegisterA := structures.NewRegister("QF2", "bool", "word", "", "",
		"dec#15F",
		"",
		map[string]string{"register_type": "word"},
	)

	validConfig.DeviceRegisters = append(validConfig.DeviceRegisters, *validRegisterA, *validRegisterB)

	// ---------------------------------------------------------------------------
	//  CASE:
	// ---------------------------------------------------------------------------
	_, err := PrepareConfig(validConfig, logger)
	if err != nil {
		t.Errorf("PrepareConfig() : Test 1 FAILED, %s", err)
	} else {
		t.Log("PrepareConfig() : Test 1 PASSED.")
	}

	validConfig.DeviceRegisters = append(validConfig.DeviceRegisters, *invalidRegisterA)

	// ---------------------------------------------------------------------------
	//  CASE:
	// ---------------------------------------------------------------------------
	_, err = PrepareConfig(validConfig, logger)
	if err != nil && err.Error() == errAwaiting.Error() {
		t.Log("PrepareConfig() : Test 2 PASSED.")
	} else {
		t.Errorf("PrepareConfig() : Test 2 FAILED, %s", err)
	}

}
