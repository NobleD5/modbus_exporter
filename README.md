# Modbus Exporter

[![codecov](https://codecov.io/gh/NobleD5/modbus_exporter/branch/main/graph/badge.svg?token=F4R3WH5VZ1)](https://codecov.io/gh/NobleD5/modbus_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/NobleD5/modbus_exporter)](https://goreportcard.com/report/github.com/NobleD5/modbus_exporter)
[![Go - Build and Test](https://github.com/NobleD5/modbus_exporter/actions/workflows/go-build-test.yml/badge.svg)](https://github.com/NobleD5/modbus_exporter/actions/workflows/go-build-test.yml)
[![Docker - Image Build and Push](https://github.com/NobleD5/modbus_exporter/actions/workflows/docker-image-push.yml/badge.svg)](https://github.com/NobleD5/modbus_exporter/actions/workflows/docker-image-push.yml)
[![CodeQL](https://github.com/NobleD5/modbus_exporter/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/NobleD5/modbus_exporter/actions/workflows/codeql-analysis.yml)

An application for collecting metrics from devices using the MODBUS protocol and converting them to the native Prometheus format.

*The exporter is not an independent data collector, it works only in integration with Prometheus. See configuration below.*

## Install
#### Docker images

```
docker pull ghcr.io/nobled5/modbus_exporter:latest
```

## Run
Binary:

```sh
./modbus_exporter --config.file=modbus.yaml
```

Other flags are available on command:

```sh
./modbus_exporter -h
```

## Modbus Config
```yaml

  DEVICE001: # must be a unique name
    device_modbus_id: 1 # default
    device_timeout: 300ms
    device_request_delay: 1000ms
    device_zero_based_addressing: false
    device_labels:
      vendor: foo
      location: data center 1
    device_registers:
      - register_name: QF1_OnOff
        register_si_name: bool
        register_type: word
        register_byte_order: big_endian
        register_address: "hex#2ee0"
        register_func_code: "FC3"
        register_labels:
          modbus_type: word
      - register_name: QF1_U_AN
        register_si_name: voltage
        register_type: uint16
        register_byte_order: big_endian
        register_address: "dec#300"
        register_func_code: "FC3"
        register_labels:
          modbus_type: uint16

  DEVICE002: # must be a unique name
    device_modbus_id: 15
    device_timeout: 300ms
    device_request_delay: 1000ms
    device_zero_based_addressing: true
    device_labels:
      vendor: bar
      location: data center 2
    device_registers:
      - register_name: QF1_I_AN
        register_si_name: current
        register_type: uint32
        register_byte_order: big_endian
        register_word_order: swapped
        register_address: "dec#300"
        register_func_code: "FC3"
        register_labels:
          modbus_type: uint32
      - register_name: QF1_I_BN
        register_si_name: voltage
        register_type: uint16
        register_byte_order: big_endian
        register_address: "dec#301"
        register_func_code: "FC3"
        register_labels:
          modbus_type: uint16
      - register_name: QF1_I_CN
        register_si_name: current
        register_type: uint16
        register_byte_order: big_endian
        register_address: "dec#302"
        register_func_code: "FC3"
        register_labels:
          modbus_type: uint16

```

## Prometheus Target Config
```yaml

  scrape_configs:

  # ----------------------------------------------------------------------

    - job_name: 'prometheus'

      static_configs:
        - targets: ['localhost:9090']

  # ----------------------------------------------------------------------

    - job_name: 'modbus-exporter'

      scrape_interval: 500ms
      scrape_timeout: 500ms

      static_configs:
        - targets: ['modbus-exporter:9700']

  # ----------------------------------------------------------------------

    - job_name: 'device'

      scrape_interval: 500ms
      scrape_timeout: 500ms

      static_configs:
        - targets: ['192.168.1.10:502']
      metrics_path: /modbus
      params:
        config:
          - DEVICE002
      relabel_configs:
        - source_labels: [__address__]
          target_label: __param_target
        - source_labels: [__param_target]
          target_label: instance
        - target_label: __address__
          replacement: modbus-exporter:9700

```
