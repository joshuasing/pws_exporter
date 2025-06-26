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

// Package homeassistant implements an exporter that publishes sensor data to
// Home Assistant via MQTT.
package homeassistant

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/joshuasing/pws_exporter/internal/exporter"
)

const (
	exporterID = "homeassistant"

	deviceClassAtmosphericPressure = "atmospheric_pressure"
	deviceClassHumidity            = "humidity"
	deviceClassPrecipitation       = "precipitation"
	deviceClassTemperature         = "temperature"
	deviceClassWindDirection       = "wind_direction"
	deviceClassWindSpeed           = "wind_speed"

	stateClassMeasurement      = "measurement"
	stateClassMeasurementAngle = "measurement_angle"
	stateClassTotalIncreasing  = "total_increasing"

	unitCelsius            = "°C"
	unitDegree             = "°"
	unitHPA                = "hPa"
	unitKPH                = "km/h"
	unitMillimeters        = "mm"
	unitPercentage         = "%"
	unitGramsPerCubicMeter = "g/m³"
)

type haComponent struct {
	ID                string
	Name              string
	Icon              string
	DeviceClass       string
	StateClass        string
	Unit              string
	ValueTemplate     string
	DisabledByDefault bool
}

var components = []haComponent{
	{
		ID:            "temperature",
		Name:          "Temperature",
		DeviceClass:   deviceClassTemperature,
		Unit:          unitCelsius,
		StateClass:    stateClassMeasurement,
		ValueTemplate: "{{ value_json.temperature }}",
	},
	{
		ID:            "humidity_abs",
		Name:          "Absolute Humidity",
		Icon:          "mdi:water-percent",
		Unit:          unitGramsPerCubicMeter,
		StateClass:    stateClassMeasurement,
		ValueTemplate: "{{ value_json.humidity_abs }}",
	},
	{
		ID:            "humidity_rel",
		Name:          "Relative Humidity",
		DeviceClass:   deviceClassHumidity,
		Unit:          unitPercentage,
		StateClass:    stateClassMeasurement,
		ValueTemplate: "{{ value_json.humidity_rel }}",
	},
	{
		ID:            "barometric_pressure",
		Name:          "Barometric Pressure",
		DeviceClass:   deviceClassAtmosphericPressure,
		Unit:          unitHPA,
		StateClass:    stateClassMeasurement,
		ValueTemplate: "{{ value_json.barometric_pressure }}",
	},
	{
		ID:            "dew_point",
		Name:          "Dew Point",
		DeviceClass:   deviceClassTemperature,
		Unit:          unitCelsius,
		StateClass:    stateClassMeasurement,
		ValueTemplate: "{{ value_json.dew_point }}",
	},
	{
		ID:            "wind_speed",
		Name:          "Wind Speed",
		DeviceClass:   deviceClassWindSpeed,
		Unit:          unitKPH,
		StateClass:    stateClassMeasurement,
		ValueTemplate: "{{ value_json.wind_speed }}",
	},
	{
		ID:            "wind_gust_speed",
		Name:          "Wind Gust",
		DeviceClass:   deviceClassWindSpeed,
		Unit:          unitKPH,
		StateClass:    stateClassMeasurement,
		ValueTemplate: "{{ value_json.wind_gust_speed }}",
	},
	{
		ID:            "wind_direction",
		Name:          "Wind Direction",
		DeviceClass:   deviceClassWindDirection,
		Unit:          unitDegree,
		StateClass:    stateClassMeasurementAngle,
		ValueTemplate: "{{ value_json.wind_direction }}",
	},
	{
		ID:            "rain_hourly",
		Name:          "Hourly Rain",
		DeviceClass:   deviceClassPrecipitation,
		Unit:          unitMillimeters,
		StateClass:    stateClassMeasurement,
		ValueTemplate: "{{ value_json.rain_past_hour }}",
	},
	{
		ID:            "rain_daily",
		Name:          "Daily Rain",
		DeviceClass:   deviceClassPrecipitation,
		Unit:          unitMillimeters,
		StateClass:    stateClassTotalIncreasing,
		ValueTemplate: "{{ value_json.rain_today }}",
	},
	{
		ID:            "indoor_temperature",
		Name:          "Indoor Temperature",
		DeviceClass:   deviceClassTemperature,
		Unit:          unitCelsius,
		StateClass:    stateClassMeasurement,
		ValueTemplate: "{{ value_json.indoor_temp }}",
	},
	{
		ID:            "indoor_humidity",
		Name:          "Indoor Relative Humidity",
		DeviceClass:   deviceClassHumidity,
		Unit:          unitPercentage,
		StateClass:    stateClassMeasurement,
		ValueTemplate: "{{ value_json.indoor_humidity }}",
	},
	{
		ID:                "feels_like",
		Name:              "Feels Like",
		DeviceClass:       deviceClassTemperature,
		Unit:              unitCelsius,
		StateClass:        stateClassMeasurement,
		ValueTemplate:     "{{ value_json.feels_like_temp }}",
		DisabledByDefault: true,
	},
	{
		ID:                "aus_apparent_temp",
		Name:              "Apparent Temperature (AUS)",
		DeviceClass:       deviceClassTemperature,
		Unit:              unitCelsius,
		StateClass:        stateClassMeasurement,
		ValueTemplate:     "{{ value_json.aus_apparent_temp }}",
		DisabledByDefault: true,
	},
	{
		ID:                "heat_index",
		Name:              "Heat Index",
		DeviceClass:       deviceClassTemperature,
		Unit:              unitCelsius,
		StateClass:        stateClassMeasurement,
		ValueTemplate:     "{{ value_json.heat_index }}",
		DisabledByDefault: true,
	},
	{
		ID:                "wind_chill",
		Name:              "Wind Chill",
		DeviceClass:       deviceClassTemperature,
		Unit:              unitCelsius,
		StateClass:        stateClassMeasurement,
		ValueTemplate:     "{{ value_json.wind_chill }}",
		DisabledByDefault: true,
	},
}

// Config is the Exporter configuration.
type Config struct {
	MQTTBrokerURL string
	MQTTUsername  string
	MQTTPassword  string
}

// Exporter is an exporter that publishes sensor information to Home Assistant
// via MQTT.
type Exporter struct {
	knownDevicesMx sync.Mutex
	knownDevices   map[string]struct{}

	mqttBrokerURL string
	mqttUsername  string
	mqttPassword  string
	mqttClient    mqtt.Client
}

var idReplacer = strings.NewReplacer("-", "_")

// NewExporter returns a new Home Assistant MQTT exporter.
func NewExporter(conf Config) (*Exporter, error) {
	return &Exporter{
		knownDevices:  make(map[string]struct{}),
		mqttBrokerURL: conf.MQTTBrokerURL,
		mqttUsername:  conf.MQTTUsername,
		mqttPassword:  conf.MQTTPassword,
	}, nil
}

// ExporterID returns the exporter ID.
func (e *Exporter) ExporterID() string {
	return exporterID
}

// HandleDeviceMeasurement handles a new measurement from a device.
func (e *Exporter) HandleDeviceMeasurement(deviceID string, dm *exporter.DeviceMeasurement) error {
	deviceID = idReplacer.Replace(deviceID)

	if err := e.publishDiscoveryIfNew(deviceID); err != nil {
		return err
	}
	return e.publishState(deviceID, dm)
}

// Run connects to the MQTT server.
func (e *Exporter) Run() error {
	if e.mqttClient != nil {
		return fmt.Errorf("already running")
	}

	slog.Info("Connecting to Home Assistant MQTT server",
		slog.String("broker_url", e.mqttBrokerURL),
		slog.String("username", e.mqttUsername))

	opts := mqtt.NewClientOptions().AddBroker(e.mqttBrokerURL)
	opts.SetUsername(e.mqttUsername)
	opts.SetPassword(e.mqttPassword)
	opts.SetAutoReconnect(true)
	opts.SetClientID("pws_exporter")

	opts.SetOnConnectHandler(func(_ mqtt.Client) {
		slog.Info("Connected to Home Assistant MQTT server",
			slog.String("broker_url", e.mqttBrokerURL),
			slog.String("username", e.mqttUsername))
	})

	e.mqttClient = mqtt.NewClient(opts)
	if token := e.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("connect to MQTT: %w", token.Error())
	}
	return nil
}

// Close disconnects from the MQTT server.
func (e *Exporter) Close() error {
	if e.mqttClient == nil {
		return nil
	}

	e.mqttClient.Disconnect(5)
	return nil
}

// publishDiscovery checks whether the device is known, and if not, then sends
// a discovery message to Home Assistant.
func (e *Exporter) publishDiscoveryIfNew(deviceID string) error {
	e.knownDevicesMx.Lock()
	if _, ok := e.knownDevices[deviceID]; ok {
		e.knownDevicesMx.Unlock()
		return nil
	}
	e.knownDevices[deviceID] = struct{}{}
	e.knownDevicesMx.Unlock()

	if err := e.publishDiscovery(deviceID); err != nil {
		// Remove from known devices so we try again.
		e.knownDevicesMx.Lock()
		delete(e.knownDevices, deviceID)
		e.knownDevicesMx.Unlock()
		return err
	}

	return nil
}

type discoveryComponent struct {
	Platform         string `json:"p"`
	Name             string `json:"name"`
	Icon             string `json:"icon,omitempty"`
	UniqueID         string `json:"unique_id"`
	DeviceClass      string `json:"device_class,omitempty"`
	StateClass       string `json:"state_class"`
	Unit             string `json:"unit_of_measurement,omitempty"`
	ValueTemplate    string `json:"value_template,omitempty"`
	EnabledByDefault bool   `json:"enabled_by_default"`
}

// publishDiscovery publishes a discovery message to Home Assistant.
func (e *Exporter) publishDiscovery(deviceID string) error {
	cmps := make(map[string]discoveryComponent, len(components))
	for _, m := range components {
		cmps[m.ID] = discoveryComponent{
			Platform:         "sensor",
			Name:             m.Name,
			Icon:             m.Icon,
			UniqueID:         fmt.Sprintf("%s_%s", deviceID, m.ID),
			DeviceClass:      m.DeviceClass,
			StateClass:       m.StateClass,
			Unit:             m.Unit,
			ValueTemplate:    m.ValueTemplate,
			EnabledByDefault: !m.DisabledByDefault,
		}
	}

	j, err := json.Marshal(map[string]any{
		"device": map[string]any{
			"ids":  deviceID,
			"name": deviceID,
		},
		"origin": map[string]any{
			"name": "pws_exporter",
			"sw":   "0.1.0", // TODO
			"url":  "https://github.com/joshuasing/pws_exporter",
		},
		"components":  cmps,
		"state_topic": fmt.Sprintf("homeassistant/device/%s/state", deviceID),
	})
	if err != nil {
		return fmt.Errorf("marshal discovery message: %w", err)
	}

	topic := fmt.Sprintf("homeassistant/device/%s/config", deviceID)
	e.mqttClient.Publish(topic, 0, true, string(j))
	slog.Debug("Published Home Assistant MQTT discovery message",
		slog.String("device_id", deviceID), slog.String("topic", topic))

	slog.Debug(string(j))

	return nil
}

// publishState publishes the sensors' states to Home Assistant.
func (e *Exporter) publishState(deviceID string, dm *exporter.DeviceMeasurement) error {
	j, err := json.Marshal(dm)
	if err != nil {
		return fmt.Errorf("marshal device measurement: %w", err)
	}

	topic := fmt.Sprintf("homeassistant/device/%s/state", deviceID)
	e.mqttClient.Publish(topic, 0, true, j)
	slog.Debug("Published sensor state to MQTT",
		slog.String("device_id", deviceID), slog.String("topic", topic))
	return nil
}

func nilEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
