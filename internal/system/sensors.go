package system

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type SensorInfo struct {
	Name  string
	Temp  float64
	Unit  string
}

type FanInfo struct {
	Name  string
	Speed float64
	Unit  string
}

type VoltageInfo struct {
	Name  string
	Value float64
	Unit  string
}

type AllSensors struct {
	Temperatures []SensorInfo
	Fans         []FanInfo
	Voltages     []VoltageInfo
}

func GetTemperatures() []SensorInfo {
	result, err := getAllSensors()
	if err != nil {
		return nil
	}
	return result.Temperatures
}

func GetFanSpeeds() []FanInfo {
	result, err := getAllSensors()
	if err != nil {
		return nil
	}
	return result.Fans
}

func GetVoltages() []VoltageInfo {
	result, err := getAllSensors()
	if err != nil {
		return nil
	}
	return result.Voltages
}

func getAllSensors() (*AllSensors, error) {
	output, err := runCmd("sensors", "-j")
	if err != nil {
		return nil, fmt.Errorf("sensors: %w", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("parse sensors output: %w", err)
	}

	result := &AllSensors{}

	for chipName, chipData := range raw {
		var attrs map[string]json.RawMessage
		if err := json.Unmarshal(chipData, &attrs); err != nil {
			continue
		}

		for key, val := range attrs {
			var inputMap map[string]json.RawMessage
			if err := json.Unmarshal(val, &inputMap); err != nil {
				continue
			}

			if strings.HasPrefix(key, "temp") {
				tempInput, ok := inputMap["temp1_input"]
				if !ok {
					continue
				}
				var temp float64
				if err := json.Unmarshal(tempInput, &temp); err != nil {
					continue
				}

				label := chipName
				if l, ok := inputMap["temp1_label"]; ok {
					var labelStr string
					if err := json.Unmarshal(l, &labelStr); err == nil {
						label = chipName + " " + labelStr
					}
				}

				result.Temperatures = append(result.Temperatures, SensorInfo{Name: label, Temp: temp, Unit: "°C"})
			}

			if strings.HasPrefix(key, "fan") {
				fanInput, ok := inputMap["fan1_input"]
				if !ok {
					continue
				}
				var speed float64
				if err := json.Unmarshal(fanInput, &speed); err != nil {
					continue
				}

				label := chipName
				if l, ok := inputMap["fan1_label"]; ok {
					var labelStr string
					if err := json.Unmarshal(l, &labelStr); err == nil {
						label = chipName + " " + labelStr
					}
				}

				result.Fans = append(result.Fans, FanInfo{Name: label, Speed: speed, Unit: "RPM"})
			}

			if strings.HasPrefix(key, "in") && key != "intrusion" {
				inInput, ok := inputMap["in0_input"]
				if !ok {
					continue
				}
				var value float64
				if err := json.Unmarshal(inInput, &value); err != nil {
					continue
				}

				label := chipName
				if l, ok := inputMap["in0_label"]; ok {
					var labelStr string
					if err := json.Unmarshal(l, &labelStr); err == nil {
						label = chipName + " " + labelStr
					}
				}

				result.Voltages = append(result.Voltages, VoltageInfo{Name: label, Value: value, Unit: "V"})
			}
		}
	}

	gpuTemp, err := getNvidiaGPUTemperature()
	if err == nil {
		result.Temperatures = append(result.Temperatures, SensorInfo{Name: "GPU", Temp: gpuTemp, Unit: "°C"})
	}

	return result, nil
}

func getNvidiaGPUTemperature() (float64, error) {
	output, err := runCmd(
		"nvidia-smi",
		"--query-gpu=temperature.gpu",
		"--format=csv,noheader",
	)
	if err != nil {
		return 0, fmt.Errorf("nvidia-smi: %w", err)
	}

	text := strings.TrimSpace(string(output))
	if text == "" {
		return 0, fmt.Errorf("empty nvidia-smi output")
	}

	temp, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("parse nvidia temperature: %w", err)
	}

	return temp, nil
}
