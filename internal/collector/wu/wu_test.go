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

package wu

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joshuasing/pws_exporter/internal/exporter"
)

const testQuery = SubmissionPath + "?ID=test&PASSWORD=testtest&action=updateraww&realtime=1&rtfreq=5&dateutc=now&baromin=29.65&tempf=63.5&dewptf=51.2&humidity=64&windspeedmph=4.4&windgustmph=4.9&winddir=270&rainin=0.0&dailyrainin=0.0&indoortempf=73.5&indoorhumidity=44"

func TestSubmission(t *testing.T) {
	var (
		deviceID        string
		lastMeasurement *exporter.DeviceMeasurement
	)
	collector := NewCollector(func(dID string, dm *exporter.DeviceMeasurement) {
		deviceID = dID
		lastMeasurement = dm
	})

	mux := http.NewServeMux()
	if err := collector.RegisterRoutes(mux); err != nil {
		t.Errorf("RegisterRoutes err = %v, want nil", err)
		return
	}
	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := ts.Client()

	// Test bad empty GET request
	res, err := client.Get(ts.URL + SubmissionPath)
	if err != nil {
		t.Errorf("bad empty request failed: %v", err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("empty request got %d, want %d", res.StatusCode, http.StatusBadRequest)
	}
	if lastMeasurement != nil {
		t.Errorf("measurement should be empty after bad request")
	}

	// Test submission
	res, err = client.Get(ts.URL + testQuery)
	if err != nil {
		t.Errorf("good submission request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("submission status got %d, want %d", res.StatusCode, http.StatusOK)
	}
	body, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		t.Errorf("failed to read response body: %v", err)
	}
	if string(body) != "success\n" {
		t.Errorf("response body got %q, want %q", string(body), "success\n")
	}
	if deviceID != "test" {
		t.Errorf("deviceID got %s, want %s", deviceID, "test")
	}
	if lastMeasurement == nil {
		t.Errorf("measurement should contain one measurement")
	}
	if lastMeasurement.Temperature != 17.5 {
		t.Errorf("temperature got %f, want %f", lastMeasurement.Temperature, 17.5)
	}
}
