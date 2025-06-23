# Personal Weather Station (PWS) exporter

[![Go Reference](https://pkg.go.dev/badge/github.com/joshuasing/pws_exporter.svg)](https://pkg.go.dev/github.com/joshuasing/pws_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/joshuasing/pws_exporter)](https://goreportcard.com/report/github.com/joshuasing/pws_exporter)
[![Go Build Status](https://github.com/joshuasing/pws_exporter/actions/workflows/go.yml/badge.svg)](https://github.com/joshuasing/pws_exporter/actions/workflows/go.yml)
[![MIT License](https://img.shields.io/badge/license-MIT-2155cc)](LICENSE)

A easy-to-use metrics exporter for off-the-shelf Personal Weather Stations (PWS).

- **Simple setup.** Start pws_exporter, configure DNS and start receiving data.
- **[Prometheus metrics](#prometheus).** Scrape metrics with Prometheus easily.
- **[Home Assistant support](#home-assistant).** Send sensor data to Home Assistant via MQTT.

**This exporter is a work in progress, things may break! If you are interested in contributing, please feel free to
contact me or create an issue/pull request!**

---

<details>
  <summary>Table of Contents</summary>

<!-- TOC -->
* [Personal Weather Station (PWS) exporter](#personal-weather-station-pws-exporter)
  * [Collectors](#collectors)
    * [DNS](#dns)
      * [Built-in server](#built-in-server)
    * [Receiving data](#receiving-data)
  * [Exporters](#exporters)
    * [Prometheus](#prometheus)
      * [Prometheus configuration](#prometheus-configuration)
    * [Home Assistant](#home-assistant)
  * [Installation](#installation)
    * [Binaries](#binaries)
    * [Docker](#docker)
  * [Contributing](#contributing)
    * [Building](#building)
    * [Contact](#contact)
      * [Security vulnerabilities](#security-vulnerabilities)
    * [License](#license)
<!-- TOC -->
</details>

## Collectors

*This project is not affiliated with Weather Underground, or any other supported services.*

pws_exporter reads data from the weather station and exposes it as metrics to Prometheus or Home Assistant.

The easiest way to receive data in most cases is by capturing the data as the weather station tries to send it to an
external API. One of the most common external submission APIs is Weather Underground, which is supported by the
majority of off-the-shelf personal weather stations.

| Name                     | URL                           | Status    |
|:-------------------------|:------------------------------|:----------|
| Weather Underground (WU) | https://www.wunderground.com/ | Supported |

I plan to add support for additional APIs in the future. If you have a weather station which supports sending data to
another API, please create an issue (or pull request) to have support added!

In the future, I would like to add support for receiving the data locally as the weather station sends it to a
controller or display. In most cases, this would require an additional device to receive the data over RF.

### DNS

Personal weather stations usually perform DNS queries to get the IP address of the external API, which allows us to
change the IP address to that of the exporter.

In order to receive data for certain collectors, the following domains must resolve to IP address of the machine running
pws_exporter.

| Domain                             | Collector Type             |
|------------------------------------|----------------------------|
| `weatherstation.wunderground.com.` | Weather Underground (`wu`) |
| `rtupdate.wunderground.com`        | Weather Underground (`wu`) |

#### Built-in server

pws_exporter has an _optional_ built-in DNS server, which can be used to change the answers of the DNS queries made by
the weather station (and return NXDOMAIN to blackhole any other queries). If used, DHCP can be configured to have the
weather station use the exporter as a DNS server.

### Receiving data

When submitting data to an external API, most personal weather stations appear to use HTTP/1.1 without TLS. Because the
standard HTTP port (`80/tcp`) will be used in most cases, the pws_exporter API server **must** receive requests on this
port in order to receive data from the weather station.

## Exporters

pws_exporter supports multiple exporter types.

- [Prometheus](#prometheus)
- [Home Assistant (MQTT)](#home-assistant)

### Prometheus

The following Prometheus metrics are exposed by this exporter. More metrics will be added soon, however some metrics may
not be supported by all APIs or weather stations.

| Metric name                                  | Description                                             |
|----------------------------------------------|---------------------------------------------------------|
| `weather_station_barometric_pressure_hpa`    | Barometric pressure in hectopascals                     |
| `weather_station_dew_point_celsius`          | Dew point in Celsius                                    |
| `weather_station_humidity_percent`           | Humidity percentage                                     |
| `weather_station_indoor_humidity`            | Indoor humidity percentage                              |
| `weather_station_indoor_temperature_celsius` | Indoor temperature in Celsius                           |
| `weather_station_rain_past_hour_mm`          | Amount of rain in the past hour in millimeters          |
| `weather_station_rain_today_mm`              | Cumulative amount of rain since midnight in millimeters |
| `weather_station_rain_today_mm`              | Cumulative amount of rain since midnight in millimeters |
| `weather_station_temperature_celsius`        | Outdoor temperature in Celsius                          |
| `weather_station_wind_direction_degrees`     | Wind direction in degrees                               |
| `weather_station_wind_gust_speed_kph`        | Wind gust speed in KM/h                                 |
| `weather_station_wind_speed_kph`             | Wind speed in KM/h                                      |

#### Prometheus configuration

To receive the metrics from pws_exporter, you need to configure Prometheus to scrape from the exporter:

```yaml
scrape_configs:
  - job_name: "pws"
    # This can be whatever you would like, but recommended to keep it lower than the
    # submission rate of the weather station so that minimal data is lost.
    scrape_interval: 5s
    static_configs:
      - targets: [ "localhost:9452" ]
```

*Change `targets` and `scrape_interval` and the address to match your setup.*

### Home Assistant

pws_exporter supports sending data to Home Assistant via MQTT. To enable this, you must set the following flags:

- `--ha-mqtt-url`: Home Assistant MQTT broker url (e.g. `tcp://homeassistant.local:1883`,
  `ssl://home.mydomain.com:8883`)
- `--ha-mqtt-user`: Home Assistant MQTT username
- `--ha-mqtt-pass`: Home Assistant MQTT password

When a pws_exporter receives data from a weather station for the first time after startup, it will automatically publish
an auto-discovery message to MQTT. When future data is received, messages will be published to the state topic.

## Installation

### Binaries

Pre-built binaries are available from [GitHub Releases](https://github.com/joshuasing/pws_exporter/releases).

You can also use `go install` to build and install a binary from source:

```shell
go install github.com/joshuasing/pws_exporter@latest
````

**Flags**

```shell
pws_exporter --help
# pws_exporter exports sensor data from Personal Weather Stations.
#
# Flags:
#       --collector-listen string   HTTP collector listen address (default ":8080")
#       --dns-listen string         DNS server listen address
#       --exporter-ip string        Exporter IP address
#       --ha-mqtt-pass string       Home Assistant MQTT broker password
#       --ha-mqtt-url string        Home Assistant MQTT broker URL
#       --ha-mqtt-user string       Home Assistant MQTT broker username
#   -h, --help                      Displays help message
#       --log string                Log level (options: debug, info, warn, error) (default "info")
#       --prom-listen string        Prometheus HTTP listen address (default ":9452")
#       --resolver string           Upstream DNS resolver address (host:port) (default "8.8.8.8:53")
#       --wu                        Whether to support Weather Underground API submission (default true)
```

**Example**

```shell
pws_exporter
# 2025/06/24 00:24:34 INFO Starting PWS exporter
# 2025/06/24 00:24:34 INFO Using Prometheus exporter
# 2025/06/24 00:24:34 INFO Using WU API collector
# 2025/06/24 00:24:34 INFO Prometheus metrics server listening address=:9452
# 2025/06/24 00:24:34 INFO Collector HTTP server listening address=:8080
```

### Docker

Docker images are published to both [GitHub Container Registry (ghcr.io)](https://ghcr.io/joshuasing/pws_exporter)
and [Docker Hub](https://hub.docker.com/r/joshuasing/pws_exporter).

```shell
docker run -p 9451:9451 -p 80:8080 ghcr.io/joshuasing/pws_exporter
# Status: Downloaded newer image for ghcr.io/joshuasing/pws_exporter:latest
# 2025/06/24 00:24:34 INFO Starting PWS exporter
# 2025/06/24 00:24:34 INFO Using Prometheus exporter
# 2025/06/24 00:24:34 INFO Using WU API collector
# 2025/06/24 00:24:34 INFO Prometheus metrics server listening address=:9452
# 2025/06/24 00:24:34 INFO Collector HTTP server listening address=:8080
```

## Contributing

All contributions are welcome! If you have found something you think could be improved, or have discovered additional
metrics you would like included, please feel free to participate by creating an issue or pull request!

### Building

Steps to build pws_exporter.

**Prerequisites**

- Go v1.23 or newer (https://go.dev/dl/)

**Build**

- Make: `make` (`make deps lint-deps` if you are missing dependencies)
- Standalone: `go build ./cmd/pws_exporter/`

### Contact

This project is maintained by Joshua Sing. You see a list of ways to contact me on my
website: https://joshuasing.dev/#contact

#### Security vulnerabilities

I take the security of my projects very seriously. As such, I strongly encourage responsible disclosure of security
vulnerabilities.

If you have discovered a security vulnerability in pws_exporter, please report it in accordance with the
project [Security Policy](SECURITY.md#reporting-a-vulnerability). **Never use GitHub issues to report a security
vulnerability.**

### License

pws_exporter is distributed under the terms of the MIT License.<br/>
For more information, please refer to the [LICENSE](LICENSE) file.
