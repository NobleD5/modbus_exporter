package collector

import (
	"fmt"
	"strings"
	"time"

	"github.com/NobleD5/modbus_exporter/pkg/master"
	"github.com/NobleD5/modbus_exporter/pkg/structures"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector structure
type Collector struct {
	Target   string
	Workload map[structures.Device]structures.Registers
	Logger   log.Logger
}

// Constants declaration section -----------------------------------------------
const (
	tInt16 = "int16"
	tInt32 = "int32"
)

// Functions declaration section -----------------------------------------------

// NewModbusCollector returns a Modbus collector ready to use.
func NewModbusCollector(
	Target string,
	Workload map[structures.Device]structures.Registers,
	Logger log.Logger,
) *Collector {
	return &Collector{
		Target:   Target,
		Workload: Workload,
		Logger:   Logger,
	}
}

// RegisterToSamples convert registers values into native prometheus metrics
func RegisterToSamples(metricType string, metricSiName string, value float64, labels map[string]string, logger log.Logger) ([]prometheus.Metric, error) {

	var (
		metricName     string
		promMetricType prometheus.ValueType
	)

	labelNames := make([]string, 0, len(labels)+1)
	labelValues := make([]string, 0, len(labels)+1)

	for k, v := range labels {
		labelNames = append(labelNames, k)
		labelValues = append(labelValues, v)
	}

	switch metricType {
	case "counter":
		promMetricType = prometheus.CounterValue
	// case "word", "dword":
	// 	promMetricType = prometheus.GaugeValue
	// case "uint16", "uint32":
	// 	promMetricType = prometheus.GaugeValue
	// case "int16", "int32":
	// 	promMetricType = prometheus.GaugeValue
	default:
		promMetricType = prometheus.GaugeValue
	}

	if metricSiName != "" {
		metricName = "modbus_" + strings.ToLower(metricSiName)
	} else {
		level.Error(logger).Log(
			"msg", "Error parsing register si name. Will use default value 'word'",
		)
		metricName = "modbus_word"
	}
	level.Debug(logger).Log(
		"metric_type", metricType,
		"metric_si_name", metricSiName,
		"metric_name", metricName,
		"metric_value", value,
		"metric_labels_names", fmt.Sprint(labelNames),
		"metric_labels_values", fmt.Sprint(labelValues),
	)

	sample, error := prometheus.NewConstMetric(
		prometheus.NewDesc(metricName, "metric.Help", labelNames, nil),
		promMetricType,
		value,
		labelValues...,
	)
	if error != nil {
		sample = prometheus.NewInvalidMetric(prometheus.NewDesc("modbus_error", "Error calling NewConstMetric", nil, nil),
			fmt.Errorf("Error for metric %s with labels %v - %s", metricName, labelValues, error.Error()))
	}

	return []prometheus.Metric{sample}, nil
}

// ScrapeTarget prepare and read workload (devices and their respective registers)
func ScrapeTarget(address string, workload map[structures.Device]structures.Registers, logger log.Logger) ([]structures.DataUnit, error) {

	var (
		error     error
		addrSlice []string
		results   []float64

		dataUnits []structures.DataUnit
		dataUnit  structures.DataUnit
	)

	addrSlice = strings.Split(address, ":")
	if len(addrSlice) < 2 {
		address = addrSlice[0] + ":502"
		level.Warn(logger).Log("msg", "Cannot find address port, using default '502'", "new_address", address)
	}

	for parameters, registers := range workload {

		results, error = master.ReadRemote(address, parameters, registers, logger)
		if error != nil {
			return nil, fmt.Errorf("Error reading remote address %s: %s", address, error.Error())
		}

		for each, register := range registers {

			dataUnit = register
			dataUnit.Value = results[each]

			level.Debug(logger).Log(
				"data_unit_value", dataUnit.Value,
				"data_unit_si_name", dataUnit.SiName,
				"data_unit_type", dataUnit.Type,
				"data_unit_byte_order", dataUnit.ByteOrder,
				"data_unit_labels", fmt.Sprint(dataUnit.Labels),
			)

			dataUnits = append(dataUnits, dataUnit)
		}

	}

	return dataUnits, nil
}

// Describe implements Prometheus.Collector
func (collector *Collector) Describe(ch chan<- *prometheus.Desc) {

	// prometheus.DescribeByCollect(collector, ch)
	// ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)

}

// Collect implements Prometheus.Collector
func (collector *Collector) Collect(ch chan<- prometheus.Metric) {

	start := time.Now()

	var (
		samples []prometheus.Metric
	)

	logger := collector.Logger

	// Target scraping
	data, error := ScrapeTarget(collector.Target, collector.Workload, logger)
	if error != nil {
		level.Error(logger).Log("msg", "ScrapeTarget() collide with an error", "target", collector.Target, "error", error.Error())
		ch <- prometheus.NewInvalidMetric(
			prometheus.NewDesc("modbus_error", "Error scraping target", nil, nil),
			error,
		)
		return
	}

	// Self metrics --------------------------------------------------------------
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("modbus_scrape_read_duration_seconds", "The time scraping target took in seconds.", nil, nil),
		prometheus.GaugeValue,
		float64(time.Since(start).Seconds()),
	)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("modbus_scrape_data_units_returned", "Data units returned from single scrape.", nil, nil),
		prometheus.GaugeValue,
		float64(len(data)),
	)

	// ---------------------------------------------------------------------------

	// Call RegisterToSamples for each received data unit after ScrapeTarget
	for _, dataUnits := range data {

		samples, error = RegisterToSamples(dataUnits.Type, dataUnits.SiName, dataUnits.Value, dataUnits.Labels, logger)
		if error != nil {
			level.Error(logger).Log("msg", "RegisterToSamples() collide with error while creating samples", "error", error.Error())
			ch <- prometheus.NewInvalidMetric(
				prometheus.NewDesc("modbus_error", "Error sampling registers", nil, nil),
				error,
			)
			return
		}
		// All ready samples to channel
		for _, sample := range samples {
			ch <- sample
		}

	}

	// Self metric ---------------------------------------------------------------
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("modbus_total_scrape_duration_seconds", "Total MODBUS time scrape took (read and processing).", nil, nil),
		prometheus.GaugeValue,
		float64(time.Since(start).Seconds()),
	)

}
