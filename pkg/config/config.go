package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gitlab.dc.miran.ru/nuzhin/modbus_exporter/pkg/structures"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gopkg.in/yaml.v2"
)

// Functions declaration section -----------------------------------------------

// Load parses the given YAML file(s) into a Config.
func Load(dirname string, logger log.Logger) (*structures.Config, error) {

	var contents []byte

	fi, error := os.Stat(dirname)
	if error != nil {
		return nil, fmt.Errorf("error while os.Stat: %s", error.Error())
	}

	switch mode := fi.Mode(); {

	case mode.IsDir():

		contents, error = LoadDirectory(dirname, logger)
		if error != nil {
			return nil, fmt.Errorf("error while LoadDirectory: %s", error.Error())
		}

	case mode.IsRegular():

		contents, error = LoadFile(dirname, logger)
		if error != nil {
			return nil, fmt.Errorf("error while LoadFile: %s", error.Error())
		}

	}

	level.Debug(logger).Log("content", fmt.Sprint(string(contents)))

	config, error := UnmarshallConf(string(contents), logger)
	if error != nil {
		return nil, fmt.Errorf("error loading read file (content) into config: %s", error.Error())
	}

	level.Debug(logger).Log("config: ", fmt.Sprint(config))

	return config, nil
}

// LoadDirectory parses the given YAML files into a slice of bytes.
func LoadDirectory(dirname string, logger log.Logger) ([]byte, error) {

	var contents []byte

	files, error := ioutil.ReadDir(dirname)
	if error != nil {
		return nil, fmt.Errorf("error reading given directory: %s", error.Error())
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("directory is empty")
	}

	level.Debug(logger).Log("files", fmt.Sprint(files))

	for _, file := range files {

		level.Debug(logger).Log("file", fmt.Sprint(file))

		path := dirname + "/" + file.Name()

		content, error := LoadFile(path, logger)
		if error != nil {
			return nil, fmt.Errorf("error reading given file (%s): %s", file.Name(), error.Error())
		}
		contents = append(contents, content...)

	}

	level.Debug(logger).Log("contents: ", fmt.Sprint(contents))

	return contents, nil
}

// LoadFile parses the given YAML file into a slice of bytes.
func LoadFile(filename string, logger log.Logger) ([]byte, error) {

	var content []byte

	content, error := ioutil.ReadFile(filename)
	if error != nil {
		return nil, fmt.Errorf("error reading given filename (%s): %s", filename, error.Error())
	}

	level.Debug(logger).Log("content: ", fmt.Sprint(string(content)))

	return content, nil
}

// UnmarshallConf decode and assigns values in the given byte slice (input) into a Config structure.
func UnmarshallConf(input string, logger log.Logger) (*structures.Config, error) {

	config := &structures.Config{}

	error := yaml.UnmarshalStrict([]byte(input), &config)
	if error != nil {
		return nil, fmt.Errorf("error unmarshaling given yaml input: %s", error.Error())
	}

	level.Debug(logger).Log("config: ", fmt.Sprint(config))

	return config, nil
}
