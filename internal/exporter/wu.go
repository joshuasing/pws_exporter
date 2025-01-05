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

package exporter

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/joshuasing/pws_exporter/internal/exporter/wu"
)

func (e *Exporter) handleWUSubmission(deviceID string, dm wu.DeviceMeasurement) {
	m := e.metrics
	l := prometheus.Labels{"station_id": deviceID}

	m.BarometricPressure.With(l).Set(float64(dm.Barometric))
	m.DewPoint.With(l).Set(float64(dm.DewPoint))
	m.Humidity.With(l).Set(float64(dm.Humidity / 100))
	m.IndoorHumidity.With(l).Set(float64(dm.IndoorHumidity / 100))
	m.IndoorTemperature.With(l).Set(float64(dm.IndoorTemp))
	m.RainPastHour.With(l).Set(float64(dm.RainPastHour))
	m.Rain.Delete(l) // Counter state is stored on the station, not in the exporter.
	m.Rain.With(l).Add(float64(dm.RainToday))
	m.Temperature.With(l).Set(float64(dm.Temperature))
	m.WindDirection.With(l).Set(float64(dm.WindDirection))
	m.WindGustSpeed.With(l).Set(float64(dm.WindGust))
	m.WindSpeed.With(l).Set(float64(dm.WindSpeed))
}
