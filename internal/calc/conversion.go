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

import "strconv"

// ParseFloat32 parses a float from the given string.
// If the string cannot be parsed as a float, 0, false will be returned.
func ParseFloat32(v string) (float32, bool) {
	f, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return 0, false
	}
	return float32(f), true
}

// FToC converts Fahrenheit to Celsius.
func FToC(f float32) float32 {
	return (f - 32) * 5 / 9
}

// CToF converts Celsius to Fahrenheit.
func CToF(c float32) float32 {
	return (c * 9 / 5) + 32
}

// InchesToMM converts inches to millimeters.
func InchesToMM(f float32) float32 {
	return f * 25.4
}

// MPHToKPH converts miles/hour to kilometers/hour.
func MPHToKPH(f float32) float32 {
	return f * 1.609344
}

// KPHToMPH converts kilometers/hour to miles/hour.
func KPHToMPH(f float32) float32 {
	return f * 0.621371
}

// KPHToMPS converts kilometers/hour to meters/second.
func KPHToMPS(f float32) float32 {
	return f / 3.6
}

// InHgToHPA converts pressure from inches of mercury (inHg) to hectopascals
// (hPa). Formula: 1 inHg = 33.8639 hPa.
func InHgToHPA(inHg float32) float32 {
	return inHg * 33.8639
}
