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
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const SubmissionPath = "/weatherstation/updateweatherstation.php"

// SubmissionAPI implements the "PWS Upload Protocol", as documented at
// https://support.weather.com/s/article/PWS-Upload-Protocol.
type SubmissionAPI struct {
	handleSubmission func(deviceID string, dm DeviceMeasurement)
}

func NewSubmissionAPI(handler func(deviceID string, dm DeviceMeasurement)) *SubmissionAPI {
	return &SubmissionAPI{
		handleSubmission: handler,
	}
}

func (wu *SubmissionAPI) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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

	slog.Info("Received WU weather data from station",
		slog.String("station_id", q.Get("ID")),
		slog.String("station_addr", remoteAddr),
		slog.String("proto", proto))

	// TODO: maybe implement password check?
	// TODO: possibly allow forwarding data to WU as well?

	var dm DeviceMeasurement
	if err := dm.fromQuery(q); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	go wu.handleSubmission(q.Get("ID"), dm)

	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "success\n")
}

// DeviceMeasurement stores sensor data submitted to the API.
type DeviceMeasurement struct {
	DateUTC      time.Time // Submission time.
	RealTime     bool      // Whether the data is real-time
	RealTimeFreq float32   // Submission frequency in seconds

	// TODO: add remaining data fields.

	WindDirection  float32 // Instantaneous wind direction, 0-360, degrees
	WindSpeed      float32 // Instantaneous wind speed, KM/h
	WindGust       float32 // Current wind gust, KM/h (software-specific time period)
	Humidity       float32 // Outdoor humidity percentage
	DewPoint       float32 // Dew point, in Celsius
	Temperature    float32 // Temperature in Celsius
	RainPastHour   float32 // Rain over past hour, millimeters
	RainToday      float32 // Rain over the past 24 hours, millimeters
	Barometric     float32 // Barometric pressure, hPA
	IndoorTemp     float32 // Indoor temperature in Celsius
	IndoorHumidity float32 // Indoor humidity, percentage
}

// fromQuery reads the measurement data from URL query values.
func (dm *DeviceMeasurement) fromQuery(q url.Values) error {
	var err error

	// Submission date
	switch q.Get("dateutc") {
	case "", "now":
		dm.DateUTC = time.Now().UTC()
	default:
		dm.DateUTC, err = time.ParseInLocation("2006-01-02 15:04:05", q.Get("dateutc"), time.UTC)
		if err != nil {
			return fmt.Errorf("parse dateutc: %w", err)
		}
	}

	// RapidFire / real-time data
	if q.Has("realtime") {
		dm.RealTime = q.Get("realtime") == "1"
	}
	if rtFreq, ok := stof(q.Get("rtfreq")); ok {
		dm.RealTimeFreq = rtFreq
	}

	// Parse data
	if windDir, ok := stof(q.Get("winddir")); ok {
		dm.WindDirection = windDir
	}
	if windSpeedMPH, ok := stof(q.Get("windspeedmph")); ok {
		dm.WindSpeed = mphToKPH(windSpeedMPH)
	}
	if windGustMPH, ok := stof(q.Get("windgustmph")); ok {
		dm.WindGust = mphToKPH(windGustMPH)
	}
	if humidity, ok := stof(q.Get("humidity")); ok {
		dm.Humidity = humidity
	}
	if dewPtf, ok := stof(q.Get("dewptf")); ok {
		dm.DewPoint = ftoc(dewPtf)
	}
	if tempf, ok := stof(q.Get("tempf")); ok {
		dm.Temperature = ftoc(tempf)
	}
	if rainIn, ok := stof(q.Get("rainin")); ok {
		dm.RainPastHour = inToMM(rainIn)
	}
	if dailyRainIn, ok := stof(q.Get("dailyrainin")); ok {
		dm.RainToday = inToMM(dailyRainIn)
	}
	if baromIn, ok := stof(q.Get("baromin")); ok {
		dm.Barometric = inHgToHPA(baromIn)
	}
	if indoorTempF, ok := stof(q.Get("indoortempf")); ok {
		dm.IndoorTemp = ftoc(indoorTempF)
	}
	if indoorHumidity, ok := stof(q.Get("indoorhumidity")); ok {
		dm.IndoorHumidity = indoorHumidity
	}

	return nil
}

// stof parses a float from the given string.
// If the string cannot be parsed as a float, 0, false will be returned.
func stof(v string) (float32, bool) {
	f, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return 0, false
	}
	return float32(f), true
}

// ftoc converts Fahrenheit to Celsius.
func ftoc(f float32) float32 {
	return (f - 32) / 1.8
}

// inToMM converts inches to millimeters.
func inToMM(f float32) float32 {
	return f * 25.4
}

// mphToKPH converts miles/hour to kilometers/hour.
func mphToKPH(f float32) float32 {
	return f * 1.609344
}

// inHgToHPA converts pressure from inches of mercury (inHg) to hectopascals
// (hPa). Formula: 1 inHg = 33.8639 hPa.
func inHgToHPA(inHg float32) float32 {
	return inHg * 33.8639
}
