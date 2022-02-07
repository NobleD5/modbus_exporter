package master

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/NobleD5/modbus_exporter/pkg/structures"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/goburrow/modbus"
)

// Types declaration section -----------------------------------------------
type (
	word  = uint16
	dword = uint32
)

// Constants declaration section -----------------------------------------------
const (
	tWord    = "word"
	tDWord   = "dword"
	tUint16  = "uint16"
	tUint32  = "uint32"
	tInt16   = "int16"
	tInt32   = "int32"
	tFloat32 = "float32"
	tFloat64 = "float64"

	bigEndian = "big_endian"
	litEndian = "lit_endian"
	swapped   = "swapped"
	mirrored  = "mirrored"

	coils            = "FC1"
	discreteInputs   = "FC2"
	holdingRegisters = "FC3"
	inputRegisters   = "FC4"
)

// Functions declaration section -----------------------------------------------

// ReadRemote return results of reading remote device registers
func ReadRemote(address string, device structures.Device, registers structures.Registers, logger log.Logger) ([]float64, error) {

	var (
		timeoutDuration      time.Duration
		requestDelayDuration time.Duration

		zeroBased bool

		readAddress, readLength uint16

		result       []byte
		finalResults []float64
	)

	timeoutDuration, error := time.ParseDuration(device.Timeout)
	if error != nil {
		level.Error(logger).Log(
			"msg", "Error parsing device timeout duration. Will use default value '200ms'",
			"err", error.Error(),
		)
		timeoutDuration = (200 * time.Millisecond)
	}

	requestDelayDuration, error = time.ParseDuration(device.RequestDelay)
	if error != nil {
		level.Error(logger).Log(
			"msg", "Error parsing device request delay duration. Will use default value '1000ms'",
			"err", error.Error(),
		)
		requestDelayDuration = (1000 * time.Millisecond)
	}

	zeroBased = device.ZeroBased
	if error != nil {
		level.Error(logger).Log(
			"msg", "Error parsing device zero-based addressing. Will use default value 'false'",
			"err", error.Error(),
		)
		zeroBased = false
	}

	level.Debug(logger).Log(
		"address", address,
		"modbus_id", device.ModbusID,
		"timeout_duration", timeoutDuration,
		"request_delay_duration", requestDelayDuration,
		"zero_based_addressing", zeroBased,
	)

	handler := modbus.NewTCPClientHandler(address)

	handler.SlaveId = byte(device.ModbusID)
	handler.Timeout = timeoutDuration
	handler.IdleTimeout = requestDelayDuration

	error = handler.Connect()
	if error != nil {
		return nil, fmt.Errorf("error connecting to address %s: %s", address, error.Error())
	}

	defer handler.Close()

	client := modbus.NewClient(handler)

	for _, register := range registers {

		for _, v := range register.Address {
			readAddress = v
		}

		switch strings.ToLower(register.Type) {
		case tWord, tUint16, tInt16:
			readLength = word(1)
			if zeroBased {
				readAddress--
			}
		case tDWord, tUint32, tInt32:
			readLength = word(2)
			if zeroBased {
				readAddress--
			}
		default:
			readLength = word(1)
			if zeroBased {
				readAddress--
			}
		}

		switch strings.ToUpper(register.FuncCode) {
		// case coils: // FC1
		// 	result, error = client.ReadCoils(readAddress, readLength)
		// case discreteInputs: // FC2
		// 	result, error = client.ReadDiscreteInputs(readAddress, readLength)
		case holdingRegisters: // FC3
			result, error = client.ReadHoldingRegisters(readAddress, readLength)
		default: // i.e. case 'inputRegisters aka FC4' or nothing
			result, error = client.ReadInputRegisters(readAddress, readLength)
		}
		if error != nil {
			// result = []byte{0}
			return nil, fmt.Errorf("modbus client read func collide into an error: %s", error)
		}

		finalResult := convertResult(
			result,
			readLength,
			register.Type,
			register.ByteOrder,
			register.WordOrder,
			logger,
		)

		level.Debug(logger).Log("final_result", finalResult)

		finalResults = append(finalResults, finalResult)

	}

	return finalResults, nil
}

func convertResult(result []byte, length word, regType, regByteOrder, regWordOrder string, logger log.Logger) float64 {

	var convertedResult float64

	regType = strings.ToLower(regType)
	regByteOrder = strings.ToLower(regByteOrder)
	regWordOrder = strings.ToLower(regWordOrder)

	level.Debug(logger).Log(
		"register_type", regType,
		"register_byte_order", regByteOrder,
		"register_word_order", regWordOrder,
	)

	if len(result) == 4 && regWordOrder != "" {

		level.Debug(logger).Log("raw_result", result, "slice_length", len(result))

		switch regWordOrder {
		case swapped:
			result = []byte{result[2], result[3], result[0], result[1]}
			level.Debug(logger).Log("swapped_result", result, "slice_length", len(result))
		case mirrored:
			result = []byte{result[3], result[2], result[1], result[0]}
			level.Debug(logger).Log("mirrored_result", result, "slice_length", len(result))
		default:
		}
	}

	switch regByteOrder {
	case litEndian:
		if regType == tUint16 || regType == tWord {
			convertedResult = float64(binary.LittleEndian.Uint16(result))
		}
		if regType == tUint32 || regType == tDWord {
			convertedResult = float64(binary.LittleEndian.Uint32(result))
		}
	default:
		if regType == tUint16 || regType == tWord {
			convertedResult = float64(binary.BigEndian.Uint16(result))
		}
		if regType == tUint32 || regType == tDWord {
			convertedResult = float64(binary.BigEndian.Uint32(result))
		}
		if regType == tInt16 {
			convertedResult = float64(int16(binary.BigEndian.Uint16(result)))
		}
		if regType == tInt32 {
			convertedResult = float64(int32(binary.BigEndian.Uint32(result)))
		}
	}

	return convertedResult
}
