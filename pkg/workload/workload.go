package workload

import (
	"strconv"
	"strings"

	"gitlab.dc.miran.ru/nuzhin/modbus_exporter/pkg/structures"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Constants declaration section -----------------------------------------------
const (
	decimalRepresentation     = "dec"
	hexadecimalRepresentation = "hex"
)

// Functions declaration section -----------------------------------------------

// PrepareConfig parses the given received config input into a valid workload
func PrepareConfig(receivedDeviceConfig *structures.Params, logger log.Logger) (map[structures.Device]structures.Registers, error) {

	var (
		// deviceTransport string
		deviceRegisters []structures.Register
		deviceLabels    map[string]string

		registerName               string
		registerRepr, registerAddr string
		rawRegister                []string

		base int
	)

	workload := make(map[structures.Device]structures.Registers)

	// deviceTransport = receivedDeviceConfig.DeviceTransport

	deviceRegisters = receivedDeviceConfig.DeviceRegisters
	deviceLabels = receivedDeviceConfig.DeviceLabels

	device := structures.NewDevice(
		receivedDeviceConfig.DeviceModbusID,
		receivedDeviceConfig.DeviceTimeout,
		receivedDeviceConfig.DeviceRequestDelay,
		receivedDeviceConfig.DeviceZeroBasedAddressing,
	)

	level.Debug(logger).Log(
		"id", device.ModbusID,
		"timeout", device.Timeout,
		"request_delay", device.RequestDelay,
	)

	// Check for key existence in a map, create and assign empty struct
	// of Registers type
	_, isPresent := workload[*device]
	if !isPresent {
		workload[*device] = structures.Registers{}
	}

	// -------------------------------------------------------------------------

	registers := []structures.DataUnit{}

	// In cycle read and prepare all registers data
	for _, register := range deviceRegisters {

		registerName = register.RegisterName

		// Split structure field RegisterAddress (eg "dec#001") in two parts
		rawRegister = strings.Split(register.RegisterAddress, "#")
		registerRepr = rawRegister[0]
		registerAddr = rawRegister[1]

		// In case which keyword (dec or hex) is used, decide which base use
		switch registerRepr {
		case decimalRepresentation:
			base = 10
		case hexadecimalRepresentation:
			base = 16
		}

		// Prepare proper modbus register address for use by 'master.ReadRemote'
		registerValue, error := strconv.ParseUint(registerAddr, base, 16)
		if error != nil {
			level.Error(logger).Log("msg", "Error parsing register address", "register_address", registerAddr, "error", error)
			return nil, error
		}

		dataUnit := structures.NewDataUnit(
			map[string]uint16{registerName: uint16(registerValue)},
			float64(0),
			register.RegisterSiName,
			register.RegisterType,
			register.RegisterByteOrder,
			register.RegisterWordOrder,
			register.RegisterFuncCode,
			register.RegisterLabels,
		)
		_, ok := dataUnit.Labels["register_name"]
		if !ok {
			dataUnit.Labels["register_name"] = register.RegisterName
		}
		// Add deviceLabels to DataUnit labels
		for k, v := range deviceLabels {
			_, ok := dataUnit.Labels[k]
			if !ok {
				dataUnit.Labels[k] = v
			}
		}

		registers = append(registers, *dataUnit)
	}

	workload[*device] = registers

	return workload, nil
}
