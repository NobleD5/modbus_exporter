package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"
)

////////////////////////////////////////////////////////////////////////////////
// Structures section for Generator.go
////////////////////////////////////////////////////////////////////////////////

// TemplConfig structure declaration
type TemplConfig map[string]TemplDevice

// TemplDevice structure declaration
type TemplDevice struct {
	Repeat   int8              `yaml:"repeat,omitempty"`
	Labels   map[string]string `yaml:"dev_labels,omitempty"`
	Sections []TemplSection    `yaml:"dev_regs_sections"`
}

// TemplSection structure declaration
type TemplSection struct {
	SectionName      string            `yaml:"section_name,omitempty"`
	SectionLabels    map[string]string `yaml:"section_labels,omitempty"`
	SectionRegisters []TemplRegisters  `yaml:"section_regs,omitempty"`
}

// TemplRegisters structure declaration
type TemplRegisters struct {
	RegistersLabels map[string]string `yaml:"regs_labels,omitempty"`
	RegistersArray  string            `yaml:"regs_array"`
}

///////////////////////////////////////////////////////////////////////////////
// Structures section for Generator.go
////////////////////////////////////////////////////////////////////////////////

// Config structure declaration
type Config map[string]Device

// Device structure declaration
type Device struct {
	Repeat   int8
	Labels   map[string]string
	Sections []Section
}

// Section structure declaration
type Section struct {
	SectionName      string
	SectionLabels    map[string]string
	SectionRegisters []Registers
}

// Registers structure declaration
type Registers struct {
	RegistersLabels map[string]string
	RegistersArray  string
}

const (
	isWord   = "word"
	isUInt16 = "uint16"
	isDWord  = "dword"
	isUInt32 = "uint32"
	isInt16  = "int16"
	isInt32  = "int32"

	firstElem  = 0
	secondElem = 1

	indent   = "\x32\x32"
	twindent = "    "
	trindent = "      "
)

////////////////////////////////////////////////////////////////////////////////
// Main Function
////////////////////////////////////////////////////////////////////////////////
func main() {
	var (
		templateFile = kingpin.Flag("template.file", "Template file which contains list of devices descriptins with modbus-registers sections.").Required().String()

		validConfig string
	)

	kingpin.CommandLine.GetFlag("help").Short('h')
	kingpin.Parse()

	logger := log.NewLogfmtLogger(os.Stdout)

	logger = level.NewFilter(logger, level.AllowError())
	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	config, error := LoadFile(logger, *templateFile)
	if error != nil {
		level.Error(logger).Log("err", error)
		os.Exit(1)
	}

	for name, device := range *config {

		for i := 0; i < int(device.Repeat); i++ {
			validConfig += name + ":\n"
			validConfig +=
				indent + "device_modbus_id: 1\n" +
					indent + "device_timeout: 200ms\n" +
					indent + "device_request_delay: 500ms\n"
			validConfig += indent + "device_labels:\n"

			for k, v := range device.Labels {
				validConfig += twindent + k + ": " + v + "\n"
			}

			validConfig += indent + "device_registers:\n"

			c, error := PrepareSections(
				logger,
				device.Sections,
			)
			if error != nil {
				level.Error(logger).Log("err", error)
			}
			validConfig += c
		}

	}

	fmt.Printf("--- config:\n%v\n\n", validConfig)

	error = ToYAML(logger, validConfig)
	if error != nil {
		level.Error(logger).Log("err", error)
	}

}

// Functions declaration section -----------------------------------------------

// Load decode and assigns values in the given byte slice (input) into a Config structure.
func Load(logger log.Logger, input string) (*TemplConfig, error) {

	config := &TemplConfig{}

	error := yaml.UnmarshalStrict([]byte(input), &config)
	if error != nil {
		level.Debug(logger).Log("stage", "strict unmarshaling given yaml input")
		return nil, error
	}

	return config, nil
}

// LoadFile parses the given YAML file into a Config.
func LoadFile(logger log.Logger, filename string) (*TemplConfig, error) {

	content, error := ioutil.ReadFile(filename)
	if error != nil {
		level.Debug(logger).Log("stage", "reading given file (*.yaml)")
		return nil, error
	}

	config, error := Load(logger, string(content))
	if error != nil {
		level.Debug(logger).Log("stage", "loading read file (content) into config")
		return nil, error
	}

	return config, nil
}

// ToYAML create predefined yaml file in current directory.
func ToYAML(logger log.Logger, safeConfig string) error {

	config, error := yaml.Marshal(safeConfig)
	if error != nil {
		level.Debug(logger).Log("stage", "marshaling safeConfig")
		return error
	}

	file, error := os.Create("./test.yaml")
	if error != nil {
		level.Error(logger).Log("msg", "error opening output file", "err", error)
	}
	config = append([]byte("# Auto-generated template config for modbus_exporter.\n"), config...)
	_, error = file.Write(config)
	if error != nil {
		level.Error(logger).Log("msg", "error writing to output file", "err", error)
	}

	return nil
}

// PrepareRegisters cyclic write and return valid configuration of registers
func PrepareRegisters(logger log.Logger, registersCount int, registersType string, labels map[string]string, firstRegister int) string {

	var validConfig string

	for i := 0; i <= registersCount; i++ {

		registerName := strconv.Itoa(firstRegister + i)

		validConfig += twindent + "- register_name: \"" + registerName + "\"\n"

		switch registersType {
		case isWord, isUInt16:
			validConfig +=
				trindent + "register_si_name: " + isUInt16 + "\n" +
					trindent + "register_type: uint16\n" +
					trindent + "register_byte_order: big_endian\n"
		case isDWord, isUInt32:
			validConfig +=
				trindent + "register_si_name: " + isUInt32 + "\n" +
					trindent + "register_type: uint32\n" +
					trindent + "register_byte_order: big_endian\n"
		case isInt16:
			validConfig +=
				trindent + "register_si_name: " + isInt16 + "\n" +
					trindent + "register_type: int16\n"
		case isInt32:
			validConfig +=
				trindent + "register_si_name: " + isInt32 + "\n" +
					trindent + "register_type: int32\n"
		}

		validConfig += trindent + "register_address: \"dec#" + registerName + "\"\n" + trindent + "register_func_code: \"FC3\"\n" + trindent + "register_labels:\n"

		for k, v := range labels {
			validConfig += twindent + twindent + k + ": " + v + "\n"
		}
	}

	return validConfig
}

// PrepareSections cyclic write and return valid configuration of sections
func PrepareSections(logger log.Logger, sections []TemplSection) (string, error) {

	var validConfig string

	for _, section := range sections {

		validConfig += "\n" + twindent + "# --- Section: " + section.SectionName + " --- #\n"

		for _, register := range section.SectionRegisters {

			registersQnty := strings.Split(register.RegistersArray, ":")
			registersArray := strings.Split(registersQnty[firstElem], "...")

			firstRegisters, error := strconv.Atoi(strings.TrimLeft(registersArray[firstElem], "{"))
			if error != nil {
				level.Error(logger).Log("err", error)
				return "", error
			}

			lastRegisters, error := strconv.Atoi(strings.TrimRight(registersArray[secondElem], "}"))
			if error != nil {
				level.Error(logger).Log("err", error)
				return "", error
			}

			var registersCount int
			if (lastRegisters - firstRegisters) > 0 {
				registersCount = lastRegisters - firstRegisters
			} else {
				registersCount = 1
			}

			var labels map[string]string

			if len(section.SectionLabels) > 0 {
				labels = section.SectionLabels
				for k, v := range register.RegistersLabels {
					labels[k] = v
				}
			} else {
				labels = map[string]string{}
				for k, v := range register.RegistersLabels {
					labels[k] = v
				}
			}

			validConfig += PrepareRegisters(
				logger,
				registersCount,
				registersQnty[secondElem],
				labels,
				firstRegisters,
			)

		}

	}

	return validConfig, nil
}

// -----------------------------------------------------------------------------

// NewDevice returns a Device ready to use.
func NewDevice(
	Repeat int8,
	Labels map[string]string,
	Sections []Section,
) *Device {
	d := &Device{
		Repeat:   Repeat,
		Labels:   Labels,
		Sections: Sections,
	}
	return d
}

// NewSection returns a Section ready to use.
func NewSection(
	SectionLabels map[string]string,
	SectionRegisters []Registers,
) *Section {
	s := &Section{
		SectionLabels:    SectionLabels,
		SectionRegisters: SectionRegisters,
	}
	return s
}

// NewRegisters returns a Regiters ready to use.
func NewRegisters(
	RegistersLabels map[string]string,
	RegistersArray string,
) *Registers {
	r := &Registers{
		RegistersLabels: RegistersLabels,
		RegistersArray:  RegistersArray,
	}
	return r
}
