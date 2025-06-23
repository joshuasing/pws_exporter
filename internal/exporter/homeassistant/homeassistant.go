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

	unitCelsius     = "°C"
	unitDegree      = "°"
	unitHPA         = "hPa"
	unitKPH         = "km/h"
	unitMillimeters = "mm"
	unitPercentage  = "%"
)

type haComponent struct {
	ID            string
	Name          string
	DeviceClass   string
	StateClass    string
	Unit          string
	ValueTemplate string
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
		ID:            "humidity",
		Name:          "Humidity",
		DeviceClass:   deviceClassHumidity,
		Unit:          unitPercentage,
		StateClass:    stateClassMeasurement,
		ValueTemplate: "{{ value_json.humidity }}",
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
		Name:          "Indoor Humidity",
		DeviceClass:   deviceClassHumidity,
		Unit:          unitPercentage,
		StateClass:    stateClassMeasurement,
		ValueTemplate: "{{ value_json.indoor_humidity }}",
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

// publishDiscovery publishes a discovery message to Home Assistant.
func (e *Exporter) publishDiscovery(deviceID string) error {
	cmps := make(map[string]any, len(components))
	for _, m := range components {
		cmps[m.ID] = map[string]any{
			"p":                   "sensor",
			"device_class":        m.DeviceClass,
			"state_class":         m.StateClass,
			"unit_of_measurement": m.Unit,
			"name":                m.Name,
			"unique_id":           fmt.Sprintf("%s_%s", deviceID, m.ID),
			"value_template":      m.ValueTemplate,
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
