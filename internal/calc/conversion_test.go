// Copyright (c) 2025 Joshua Sing <joshua@Joshuasing.dev>
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

package calc

import (
	"testing"
)

func TestFtoC(t *testing.T) {
	tts := []struct {
		F float64
		C float64
	}{
		{F: -9.4, C: -23},     // Extremely cold
		{F: 0.0, C: -17.7778}, // Below freezing
		{F: 32, C: 0},         // Freezing point of water
		{F: 77.50, C: 25.27},  // Average temperature in Summer for Melbourne, AU
		{F: 98.6, C: 37},      // Normal body temperature
		{F: 212, C: 100},      // Boiling point of water
	}
	for _, tt := range tts {
		if got := FToC(tt.F); !approxEqual(got, tt.C) {
			t.Errorf("FToC(%f) = %f, want %f", tt.F, got, tt.C)
		}
	}
}

func TestCToF(t *testing.T) {
	tts := []struct {
		F float64
		C float64
	}{
		{C: -23, F: -9.4},     // Extremely cold
		{C: -17.7778, F: 0.0}, // Below freezing
		{C: 0, F: 32},         // Freezing point of water
		{C: 25.30, F: 77.54},  // Average temperature in Summer for Melbourne, AU
		{C: 37, F: 98.6},      // Normal body temperature
		{C: 100, F: 212},      // Boiling point of water
	}
	for _, tt := range tts {
		if got := CToF(tt.C); !approxEqual(got, tt.F) {
			t.Errorf("CToF(%f) = %f, want %f", tt.C, got, tt.F)
		}
	}
}

func TestInchesToMM(t *testing.T) {
	tts := []struct {
		In float64
		Mm float64
	}{
		{In: 0, Mm: 0},
		{In: 1, Mm: 25.4},
		{In: 2.5, Mm: 63.5},
		{In: 10, Mm: 254},
	}
	for _, tt := range tts {
		if got := InchesToMM(tt.In); !approxEqual(got, tt.Mm) {
			t.Errorf("InchesToMM(%f) = %f, want %f", tt.In, got, tt.Mm)
		}
	}
}

func TestMPHToKPH(t *testing.T) {
	tts := []struct {
		MPH float64
		KPH float64
	}{
		{MPH: 0, KPH: 0},
		{MPH: 1, KPH: 1.60934},
		{MPH: 45, KPH: 72.4203},
		{MPH: 60, KPH: 96.56064},
	}
	for _, tt := range tts {
		if got := MPHToKPH(tt.MPH); !approxEqual(got, tt.KPH) {
			t.Errorf("MPHToKPH(%f) = %f, want %f", tt.MPH, got, tt.KPH)
		}
	}
}

func TestKPHToMPS(t *testing.T) {
	tts := []struct {
		KPH float64
		MPS float64
	}{
		{KPH: 0, MPS: 0},
		{KPH: 1, MPS: 0.277},
		{KPH: 7.2, MPS: 2},
		{KPH: 50, MPS: 13.8889},
	}
	for _, tt := range tts {
		if got := KPHToMPS(tt.KPH); !approxEqual(got, tt.MPS) {
			t.Errorf("KPHToMPS(%f) = %f, want %f", tt.KPH, got, tt.MPS)
		}
	}
}

func TestInHGToHPA(t *testing.T) {
	tts := []struct {
		InHG float64
		HPA  float64
	}{
		{InHG: 0, HPA: 0},
		{InHG: 1, HPA: 33.8639},
		{InHG: 5, HPA: 169.3195},
		{InHG: 29.92, HPA: 1013.207},
	}
	for _, tt := range tts {
		if got := InHgToHPA(tt.InHG); !approxEqual(got, tt.HPA) {
			t.Errorf("InHgToHPA(%f) = %f, want %f", tt.InHG, got, tt.HPA)
		}
	}
}
