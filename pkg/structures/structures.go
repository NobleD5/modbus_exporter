package structures

import (
	"sync"
)

// Types declaration section ---------------------------------------------------

////////////////////////////////////////////////////////////////////////////////
// Structures section for main.go
////////////////////////////////////////////////////////////////////////////////

// SafeConfig structure declaration
type SafeConfig struct {
	sync.RWMutex
	C *Config
}

////////////////////////////////////////////////////////////////////////////////
// Structures section for config.go
////////////////////////////////////////////////////////////////////////////////

// Config structure declaration
type Config map[string]*Params

// Params structure declaration
type Params struct {
	DeviceTransport           string            `yaml:"device_transport,omitempty"`
	DeviceTimeout             string            `yaml:"device_timeout,omitempty"`
	DeviceModbusID            byte              `yaml:"device_modbus_id,omitempty"`
	DeviceZeroBasedAddressing bool              `yaml:"device_zero_based_addressing,omitempty"`
	DeviceRequestDelay        string            `yaml:"device_request_delay,omitempty"`
	DeviceLabels              map[string]string `yaml:"device_labels,omitempty"`
	DeviceRegisters           []Register        `yaml:"device_registers"`
}

// Register structure declaration
type Register struct {
	RegisterName      string            `yaml:"register_name"`
	RegisterSiName    string            `yaml:"register_si_name,omitempty"`
	RegisterType      string            `yaml:"register_type,omitempty"`
	RegisterByteOrder string            `yaml:"register_byte_order,omitempty"`
	RegisterWordOrder string            `yaml:"register_word_order,omitempty"`
	RegisterAddress   string            `yaml:"register_address"`
	RegisterFuncCode  string            `yaml:"register_func_code,omitempty"`
	RegisterLabels    map[string]string `yaml:"register_labels,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// Structures section for workload.go
////////////////////////////////////////////////////////////////////////////////

// Device structure declaration
type Device struct {
	Timeout, RequestDelay string
	ModbusID              byte
	ZeroBased             bool
}

// Registers structure declaration
type Registers []DataUnit

// DataUnit structure declaration
type DataUnit struct {
	Address                                      map[string]uint16
	Value                                        float64
	SiName, Type, ByteOrder, WordOrder, FuncCode string
	Labels                                       map[string]string
}

// Functions declaration section -----------------------------------------------

// NewRegister returns a Register ready to use.
func NewRegister(
	RegisterName string,
	RegisterSiName string,
	RegisterType string,
	RegisterByteOrder string,
	RegisterWordOrder string,
	RegisterAddress string,
	RegisterFuncCode string,
	RegisterLabels map[string]string,
) *Register {
	r := &Register{
		RegisterName:      RegisterName,
		RegisterSiName:    RegisterSiName,
		RegisterType:      RegisterType,
		RegisterByteOrder: RegisterByteOrder,
		RegisterWordOrder: RegisterWordOrder,
		RegisterAddress:   RegisterAddress,
		RegisterFuncCode:  RegisterFuncCode,
		RegisterLabels:    RegisterLabels,
	}
	return r
}

// NewDevice returns a Device ready to use.
func NewDevice(
	ModbusID byte,
	Timeout string,
	RequestDelay string,
	ZeroBased bool,
) *Device {
	d := &Device{
		ModbusID:     ModbusID,
		Timeout:      Timeout,
		RequestDelay: RequestDelay,
		ZeroBased:    ZeroBased,
	}
	return d
}

// NewDataUnit returns a DataUnit ready to use.
func NewDataUnit(
	Address map[string]uint16,
	Value float64,
	SiName string,
	Type string,
	ByteOrder string,
	WordOrder string,
	FuncCode string,
	Labels map[string]string,
) *DataUnit {
	du := &DataUnit{
		Address:   Address,
		Value:     Value,
		SiName:    SiName,
		Type:      Type,
		ByteOrder: ByteOrder,
		WordOrder: WordOrder,
		FuncCode:  FuncCode,
		Labels:    Labels,
	}
	return du
}

// NewSafeConfig return a SafeConfig ready to use.
func NewSafeConfig(
	C *Config,
) *SafeConfig {
	sc := &SafeConfig{
		C: C,
	}
	return sc
}

// safeConfig = &SafeConfig{
// 	C: &structures.Config{},
// }
