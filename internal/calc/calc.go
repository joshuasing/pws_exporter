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

// Package calc provides utilities for calculating additional metrics, as well
// as converting between data formats.
package calc

import (
	"errors"
	"math"
)

// AbsoluteHumidity returns the absolute humidity in grams per cubic meter
// (g/m³) for a given temperature in Celsius and relative humidity (0.0-1.0).
func AbsoluteHumidity(tempC, relHumidity float64) (float64, error) {
	if tempC < -273.15 {
		return 0, errors.New("temp cannot be below absolute zero")
	}
	if relHumidity < 0 || relHumidity > 1 {
		return 0, errors.New("relative humidity must be between 0.0 and 1.0")
	}

	// Saturation vapor pressure (hPa) with Magnus formula
	es := 6.112 * math.Exp((17.67*tempC)/(tempC+243.5))

	// Actual vapor pressure (hPa)
	ea := relHumidity * es

	// Absolute humidity (g/m³)
	return (ea * 216.7) / (tempC + 273.15), nil
}

// DewPoint calculates the dew point temperature in Celsius, using the
// Magnus-Tetens approximation.
//
// References:
//   - https://iridl.ldeo.columbia.edu/dochelp/QA/Basic/dewpoint.html
//   - https://www.wpc.ncep.noaa.gov/html/dewrh.shtml
func DewPoint(tempC, relHumidity float64) (float64, error) {
	if relHumidity <= 0 || relHumidity > 1 {
		return 0, errors.New("relative humidity must be between 0.0 and 1.0")
	}

	alpha := math.Log(relHumidity) + (17.625*tempC)/(243.04+tempC)
	dewPoint := (243.04 * alpha) / (17.625 - alpha)

	return dewPoint, nil
}

// FeelsLike returns the apparent ("feels like") temperature in Celcius,
// based on NOAA approach combining heat index and wind chill.
//
// For temperatures >= 26.7°C (80°F), it uses the Heat Index to account for
// effects caused by humidity.
//
// For temperatures <= 10°C (50°F) and wind speeds >= ~1.34 m/s (3 mph), it uses
// Wind Chill to account for cooling effects caused by wind.
//
// For temperatures outside of these ranges, it returns the air temperature.
//
// References:
//   - https://www.wpc.ncep.noaa.gov/html/heatindex_equation.shtml
//   - https://www.weather.gov/media/epz/wxcalc/windChill.pdf
func FeelsLike(tempC, relHumidity, windSpeedKPH float64) (float64, error) {
	if relHumidity < 0 || relHumidity > 1 {
		return 0, errors.New("relative humidity must be between 0.0 and 1.0")
	}
	if windSpeedKPH < 0 {
		return 0, errors.New("wind speed cannot be negative")
	}

	switch {
	case tempC >= 26.7: // 80°F
		return HeatIndex(tempC, relHumidity)
	case tempC <= 10: // 50°F
		return WindChill(tempC, windSpeedKPH)
	default:
		return tempC, nil
	}
}

// ApparentTemperatureAU calculates the apparent ("feels like") temperature in
// Celsius, using the Australian Bureau of Meteorology (BoM) standard, based on
// the Steadman Apparent Temperature.
//
// It considers air temperature (°C), relative humidity, wind speed (km/s), and
// optionally net radiation absorbed per unit area of body surface (W/m²).
//
// References:
//   - https://www.bom.gov.au/info/thermal_stress/
//   - https://web.archive.org/web/20240823111037/http://www.bom.gov.au/jshess/docs/1994/steadman.pdf
func ApparentTemperatureAU(tempC, relHumidity, windSpeedKPH, solarRadiation float64) (float64, error) {
	if relHumidity < 0 || relHumidity > 1 {
		return 0, errors.New("relative humidity must be between 0.0 and 1.0")
	}
	if windSpeedKPH < 0 {
		return 0, errors.New("wind speed cannot be negative")
	}
	if solarRadiation < 0 {
		return 0, errors.New("solar radiation cannot be negative")
	}

	ws := KPHToMPS(windSpeedKPH) // convert to meters/second

	// Vapour pressure in hPa from RH and temperature (BoM formula)
	// e = RH × 6.105 × exp(17.27 × T / (237.7 + T))
	e := relHumidity * 6.105 * math.Exp(17.27*tempC/(237.7+tempC))

	// Apparent temperature base (without solar radiation)
	var at float64
	switch {
	case solarRadiation > 0:
		// Formula including the effects of temperature, humidity, wind and
		// solar radiation.
		at = tempC + 0.348*e - 0.7*ws + 0.7*(solarRadiation)/(ws+10) - 4.25
	default:
		// Formula including the effects of temperature, humidity and wind.
		at = tempC + 0.33*e - 0.70*ws - 4.0
	}

	return at, nil
}

// HeatIndex calculates the Heat Index in Celcius, based on NOAA standard.
//
// The formula only applies if the temperature is above 26.7°C (~80°F).
//
// References:
//   - https://en.wikipedia.org/wiki/Heat_index
//   - https://www.wpc.ncep.noaa.gov/html/heatindex_equation.shtml
func HeatIndex(tempC, relHumidity float64) (float64, error) {
	if relHumidity < 0 || relHumidity > 1 {
		return 0, errors.New("relative humidity must be between 0 and 1")
	}
	if tempC < 26.6 {
		return tempC, nil
	}

	// Rothfusz regression constants
	const (
		c1 = -42.379
		c2 = 2.04901523
		c3 = 10.14333127
		c4 = -0.22475541
		c5 = -6.83783e-3
		c6 = -5.481717e-2
		c7 = 1.22874e-3
		c8 = 8.5282e-4
		c9 = -1.99e-6
	)

	tempF := CToF(tempC)
	rh := relHumidity * 100 // convert to percent

	// Rothfusz regression
	heatIndex := c1 +
		c2*tempF +
		c3*rh +
		c4*tempF*rh +
		c5*tempF*tempF +
		c6*rh*rh +
		c7*tempF*tempF*rh +
		c8*tempF*rh*rh +
		c9*tempF*tempF*rh*rh

	// Adjustments
	switch {
	case rh < 13 && tempF >= 80 && tempF <= 112:
		heatIndex -= ((13 - rh) / 4) * math.Sqrt((17-math.Abs(tempF-95))/17)
	case rh > 85 && tempF >= 80 && tempF >= 87:
		heatIndex += ((rh - 85) / 10) * ((87 - tempF) / 5)
	}

	return FToC(heatIndex), nil
}

// WindChill calculates the Wind Chill in Celcius, based on NOAA standard.
//
// The formula only applies if wind speed >= 4.9 km/h (~3 mph) and temperature
// is below 10°C (50°F).
//
// References:
//   - https://en.wikipedia.org/wiki/Wind_chill
//   - https://www.weather.gov/media/epz/wxcalc/windChill.pdf
func WindChill(tempC, windSpeedKPH float64) (float64, error) {
	if windSpeedKPH < 0 {
		return 0, errors.New("wind speed cannot be negative")
	}
	if windSpeedKPH < 4.8 || tempC > 10 {
		return tempC, nil
	}

	tempF := CToF(tempC)
	windSpeedMPH := KPHToMPH(windSpeedKPH)

	windChillF := 35.74 + (0.6215 * tempF) -
		(35.75 * math.Pow(windSpeedMPH, 0.16)) +
		(0.4275 * tempF * math.Pow(windSpeedMPH, 0.16))

	return FToC(windChillF), nil
}
