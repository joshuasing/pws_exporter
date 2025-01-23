# Personal Weather Station (PWS) exporter

[![Go Reference](https://pkg.go.dev/badge/github.com/joshuasing/pws_exporter.svg)](https://pkg.go.dev/github.com/joshuasing/pws_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/joshuasing/pws_exporter)](https://goreportcard.com/report/github.com/joshuasing/pws_exporter)
[![Go Build Status](https://github.com/joshuasing/pws_exporter/actions/workflows/go.yml/badge.svg)](https://github.com/joshuasing/pws_exporter/actions/workflows/go.yml)
[![MIT License](https://img.shields.io/badge/license-MIT-2155cc)](LICENSE)

A Prometheus Exporter for off-the-shelf Personal Weather Stations (PWS).<br/>
*This project is not affiliated with Weather Underground, or any other supported submission APIs.*

**This exporter is a work in progress, things may break! If you are interested in contributing, please feel free to
contact me or create an issue/pull request!**

## Supported submission APIs

pws_exporter can capture data from the weather station when it tries to send it to an external API. One of the most
common external submission APIs is Weather Underground, which is supported by the majority of off-the-shelf personal
weather stations.

Currently, pws_exporter only supports Weather Underground, however I plan to add support for other APIs in the future.
If you have a weather station which supports sending data to another API, please create an issue (or pull request) to
have support added!

| Name                     | URL                           | Status    |
|:-------------------------|:------------------------------|:----------|
| Weather Underground (WU) | https://www.wunderground.com/ | Supported |

### DNS

Personal weather stations usually perform DNS queries to get the IP address of the external API, which allows us to
change the IP address to that of the exporter.

pws_exporter has an optional built-in DNS server, which can be used to change the answers of the DNS queries made by the
weather station (and return NXDOMAIN to blackhole any other queries). If used, DHCP can be configured to have the
weather station use the exporter as a DNS server.

### Receiving data

When submitting data to an external API, most personal weather stations appear to use HTTP/1.1 without TLS. Because the
standard HTTP port (`80/tcp`) will be used in most cases, the pws_exporter API server must be listening on this port in
order to receive data from the weather station.

In cases where TLS is used, manufacturers may opt to have the device skip verifying the server's TLS certificate to
remove the need of having root CA certificates on the device. This means that pws_exporter may be able to still
intercept traffic by listening on port `443/tcp` and using a self-signed TLS certificate.

## Metrics

The following metrics are exposed by this exporter. More metrics will be added soon, however some metrics may not be
supported by all APIs or weather stations.

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
| `weather_station_wind_gust_kph`              | Wind gust speed in KM/h                                 |
| `weather_station_wind_speed_kph`             | Wind speed in KM/h                                      |

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
# Usage of pws_exporter:
#  -dns-listen string
#        DNS server listen address
#  -exporter string
#        Exporter IP address
#  -listen string
#        Listen address (default ":9452")
#  -log string
#        Log level (default "info")
#  -resolver string
#        Upstream DNS resolver (default "8.8.8.8:53")
#  -wu-listen string
#        WU HTTP server listen address (default ":80")
#  -wu-tls-listen string
#        WU HTTPS server listen address (default ":443")
```

**Example**

```shell
pws_exporter
# 2025/01/23 23:09:18 INFO Starting WU Weather Station exporter
# 2025/01/23 23:09:18 INFO Metrics HTTP server listening address=:9452
# 2025/01/23 23:09:18 INFO WU API TLS server listening address=[::]:443
# 2025/01/23 23:09:18 INFO WU API server listening address=:80
```

### Docker

Docker images are published to both [GitHub Container Registry (ghcr.io)](https://ghcr.io/joshuasing/pws_exporter)
and [Docker Hub](https://hub.docker.com/r/joshuasing/pws_exporter).

```shell
docker run -p 9451:9451 ghcr.io/joshuasing/pws_exporter:latest
# Status: Downloaded newer image for ghcr.io/joshuasing/pws_exporter:latest
# 2025/01/23 23:09:18 INFO Starting WU Weather Station exporter
# 2025/01/23 23:09:18 INFO Metrics HTTP server listening address=:9452
# 2025/01/23 23:09:18 INFO WU API TLS server listening address=[::]:443
# 2025/01/23 23:09:18 INFO WU API server listening address=:80
```

### Prometheus

To use the PWS Prometheus Exporter, you need to configure Prometheus to scrape from the exporter:

```yaml
scrape_configs:
  - job_name: "pws"
    # This can be whatever you would like, but recommended to keep it lower than the
    # submission rate of the weather station so that minimal data is lost.
    scrape_interval: 5s
    static_configs:
      - targets: [ "localhost:9452" ]
```

*Change `scrape_interval` and the address to match your setup.*

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
