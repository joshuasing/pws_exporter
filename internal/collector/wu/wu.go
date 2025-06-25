// Copyright (c) 2025 Joshua Sing <joshua@joshuasing.dev>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package wu implements the Weather Underground data submission API, as
// specified by https://support.weather.com/s/article/PWS-Upload-Protocol.
//
// The submitted data uses imperial values, which are immediately converted to
// metric values for compatibility with other systems.
package wu

import (
	"crypto/tls"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/joshuasing/pws_exporter/internal/calc"
	"github.com/joshuasing/pws_exporter/internal/collector"
	"github.com/joshuasing/pws_exporter/internal/exporter"
)

// SubmissionPath is the path used to submit data.
const SubmissionPath = "/weatherstation/updateweatherstation.php"

// wuDomains are domains used to submit data to the Weather Underground (WU)
// API. These domains must resolve to the IP address of the collector in order
// for data to be collected.
var wuDomains = []string{
	"weatherstation.wunderground.com", // Standard submission API
	"rtupdate.wunderground.com",       // RapidFire (real-time) submission API
}

type wuParser func(v string, dm *exporter.DeviceMeasurement) error

var wuMappings = map[string]wuParser{
	"dateutc": func(v string, dm *exporter.DeviceMeasurement) error {
		if v == "" || v == "now" {
			dm.LastUpdated = time.Now().UTC()
			return nil
		}

		t, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.UTC)
		if err != nil {
			return err
		}
		dm.LastUpdated = t
		return nil
	},
	"rtfreq": func(v string, dm *exporter.DeviceMeasurement) error {
		if rtFreq, ok := calc.ParseFloat32(v); ok {
			dm.UpdateFreq = rtFreq
		}
		return nil
	},
	"winddir": func(v string, dm *exporter.DeviceMeasurement) error {
		if windDir, ok := calc.ParseFloat32(v); ok {
			dm.WindDirection = windDir
		}
		return nil
	},
	"windspeedmph": func(v string, dm *exporter.DeviceMeasurement) error {
		if windSpeedMPH, ok := calc.ParseFloat32(v); ok {
			dm.WindSpeed = calc.MPHToKPH(windSpeedMPH)
		}
		return nil
	},
	"windgustmph": func(v string, dm *exporter.DeviceMeasurement) error {
		if windGustSpeedMPH, ok := calc.ParseFloat32(v); ok {
			dm.WindGustSpeed = calc.MPHToKPH(windGustSpeedMPH)
		}
		return nil
	},
	"humidity": func(v string, dm *exporter.DeviceMeasurement) error {
		if humidity, ok := calc.ParseFloat32(v); ok {
			dm.RelativeHumidity = humidity / 100
		}
		return nil
	},
	"dewptf": func(v string, dm *exporter.DeviceMeasurement) error {
		if dewPointF, ok := calc.ParseFloat32(v); ok {
			dm.DewPoint = calc.FToC(dewPointF)
		}
		return nil
	},
	"tempf": func(v string, dm *exporter.DeviceMeasurement) error {
		if tempF, ok := calc.ParseFloat32(v); ok {
			dm.Temperature = calc.FToC(tempF)
		}
		return nil
	},
	"rainin": func(v string, dm *exporter.DeviceMeasurement) error {
		if rainPastHourIn, ok := calc.ParseFloat32(v); ok {
			dm.RainPastHour = calc.InchesToMM(rainPastHourIn)
		}
		return nil
	},
	"dailyrainin": func(v string, dm *exporter.DeviceMeasurement) error {
		if dailyRainIn, ok := calc.ParseFloat32(v); ok {
			dm.RainToday = calc.InchesToMM(dailyRainIn)
		}
		return nil
	},
	"baromin": func(v string, dm *exporter.DeviceMeasurement) error {
		if barometricPressureIn, ok := calc.ParseFloat32(v); ok {
			dm.BarometricPressure = calc.InHgToHPA(barometricPressureIn)
		}
		return nil
	},
	"indoortempf": func(v string, dm *exporter.DeviceMeasurement) error {
		if indoorTempF, ok := calc.ParseFloat32(v); ok {
			dm.IndoorTemp = calc.FToC(indoorTempF)
		}
		return nil
	},
	"indoorhumidity": func(v string, dm *exporter.DeviceMeasurement) error {
		if indoorHumidity, ok := calc.ParseFloat32(v); ok {
			dm.IndoorHumidity = indoorHumidity / 100
		}
		return nil
	},
}

// Collector collects data being sent using the "PWS Upload Protocol", as
// documented at https://support.weather.com/s/article/PWS-Upload-Protocol.
type Collector struct {
	handler collector.HandlerFunc
}

// NewCollector returns a new WU collector.
func NewCollector(handler collector.HandlerFunc) *Collector {
	return &Collector{
		handler: handler,
	}
}

// Domains returns domains used to submit data to WU.
func (wu *Collector) Domains() []string {
	return wuDomains
}

// RegisterRoutes registers the HTTP routes used to submit data.
func (wu *Collector) RegisterRoutes(mux *http.ServeMux) error {
	mux.Handle("GET "+SubmissionPath, http.HandlerFunc(wu.handleSubmission))
	return nil
}

func (wu *Collector) handleSubmission(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	if q.Get("action") != "updateraww" || !q.Has("ID") || !q.Has("PASSWORD") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	proto := req.Proto
	if req.TLS != nil {
		proto += " " + tls.VersionName(req.TLS.Version)
	}
	remoteAddr, _, _ := net.SplitHostPort(req.RemoteAddr)

	stationID := q.Get("ID")
	slog.Info("Received WU weather data from station",
		slog.String("station_id", stationID),
		slog.String("station_addr", remoteAddr),
		slog.String("proto", proto))

	// TODO: maybe implement password check?
	// TODO: possibly allow forwarding data to WU as well?

	// Parse data
	dm := new(exporter.DeviceMeasurement)
	for k, handle := range wuMappings {
		if err := handle(q.Get(k), dm); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}

	// Derive additional metrics
	exporter.DeriveMetrics(dm)

	go wu.handler(stationID, dm)

	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "success\n")
}
