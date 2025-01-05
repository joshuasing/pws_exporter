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
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testQuery = SubmissionPath + "?ID=test&PASSWORD=testtest&action=updateraww&realtime=1&rtfreq=5&dateutc=now&baromin=29.65&tempf=63.5&dewptf=51.2&humidity=64&windspeedmph=4.4&windgustmph=4.9&winddir=270&rainin=0.0&dailyrainin=0.0&indoortempf=73.5&indoorhumidity=44"

func TestSubmission(t *testing.T) {
	var (
		stationID       string
		lastMeasurement *DeviceMeasurement
	)
	sapi := NewSubmissionAPI(func(sID string, dm DeviceMeasurement) {
		stationID = sID
		lastMeasurement = &dm
	})

	ts := httptest.NewTLSServer(sapi)
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
	if stationID != "test" {
		t.Errorf("stationID got %s, want %s", stationID, "test")
	}
	if lastMeasurement == nil {
		t.Errorf("measurement should contain one measurement")
	}
	if lastMeasurement.Temperature != 17.5 {
		t.Errorf("temperature got %f, want %f", lastMeasurement.Temperature, 17.5)
	}
}

func TestFtoC(t *testing.T) {
	tts := []struct {
		F float32
		C float32
	}{
		{F: 0.0, C: -17.7778}, // Below freezing
		{F: 32, C: 0},         // Freezing point of water
		{F: 77.5, C: 25.3},    // Average temperature in Summer for Melbourne, AU
		{F: 98.6, C: 37},      // Normal body temperature
		{F: 212, C: 100},      // Boiling point of water
	}
	for _, tt := range tts {
		if c := ftoc(tt.F); round(c, 4) != round(c, 4) {
			t.Errorf("ftoc(%f) = %f, want %f", tt.F, c, tt.C)
		}
	}
}

func TestInToMM(t *testing.T) {
	tts := []struct {
		In float32
		Mm float32
	}{
		{In: 0, Mm: 0},
		{In: 1, Mm: 25.4},
		{In: 2.5, Mm: 63.5},
		{In: 10, Mm: 254},
	}
	for _, tt := range tts {
		if mm := inToMM(tt.In); round(mm, 4) != round(tt.Mm, 4) {
			t.Errorf("intomm(%f) = %f, want %f", tt.In, mm, tt.Mm)
		}
	}
}

func TestMPHToKPH(t *testing.T) {
	tts := []struct {
		MPH float32
		KPH float32
	}{
		{MPH: 0, KPH: 0},
		{MPH: 1, KPH: 1.60934},
		{MPH: 45, KPH: 72.4203},
		{MPH: 60, KPH: 96.56064},
	}
	for _, tt := range tts {
		if c := mphToKPH(tt.MPH); round(c, 5) != round(c, 5) {
			t.Errorf("mphToKPH(%f) = %f, want %f", tt.MPH, c, tt.KPH)
		}
	}
}

func TestInHGToHPA(t *testing.T) {
	tts := []struct {
		InHG float32
		HPA  float32
	}{
		{InHG: 0, HPA: 0},
		{InHG: 1, HPA: 33.8639},
		{InHG: 5, HPA: 169.3195},
		{InHG: 29.92, HPA: 1013.25},
	}
	for _, tt := range tts {
		if c := inHgToHPA(tt.InHG); round(c, 4) != round(c, 4) {
			t.Errorf("inHgToHPA(%f) = %f, want %f", tt.InHG, c, tt.HPA)
		}
	}
}

func round(v float32, places int) float32 {
	factor := math.Pow(10, float64(places))
	return float32(math.Round(float64(v)*factor) / factor)
}
