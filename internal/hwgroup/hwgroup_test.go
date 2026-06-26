package hwgroup

import (
	"testing"

	"github.com/gosnmp/gosnmp"
)

func TestMapSNMPVersion(t *testing.T) {
	tests := []struct {
		input       string
		expected    gosnmp.SnmpVersion
		expectedErr bool
	}{
		{"1", gosnmp.Version1, false},
		{"2c", gosnmp.Version2c, false},
		{"3", gosnmp.Version3, false},
		{"invalid", gosnmp.Version1, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual, err := MapSNMPVersion(tt.input)

			if (err != nil) != tt.expectedErr {
				t.Errorf("actual: %v, expected: %v", err, tt.expectedErr)
				return
			}

			if actual != tt.expected {
				t.Errorf("actual: %v, expected: %v", actual, tt.expected)
			}
		})
	}
}

func TestIsSupportedDevice(t *testing.T) {
	tests := []struct {
		input          string
		expectedDevice string
		expectedOK     bool
	}{
		{"Poseidon-1", "Poseidon", true},
		{"Damocles Unit", "Damocles", true},
		{"STE2 Sensor", "STE2", true},
		{"WLD Box", "WLD", true},
		{"Unknown Device", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actualDev, actualOK := IsSupportedDevice(tt.input)

			if actualOK != tt.expectedOK {
				t.Errorf("actual: %v, expected: %v", actualOK, tt.expectedOK)
			}

			if actualDev != tt.expectedDevice {
				t.Errorf("actual: %v, expected: %v", actualDev, tt.expectedDevice)
			}
		})
	}
}

func TestContactOutputFields_ContactString(t *testing.T) {
	cf := ContactOutputFields{Name: "Unit", AlarmState: 0, AlarmSetup: 0}

	expected := "Contact name: Unit, AlarmState: normal, AlarmSetup: active if on"

	if actual := cf.ContactString(); actual != expected {
		t.Errorf("actual: %v, expected: %v", actual, expected)
	}
}

func TestContactOutputFields_OutputString(t *testing.T) {
	cf := ContactOutputFields{Name: "Unit", AlarmState: 2, AlarmSetup: 3}

	expected := "Output name: Unit, Type: On (+10V) / Off (0V) (DTR output), Mode: On if value higher than Trigger"

	if actual := cf.OutputString(); actual != expected {
		t.Errorf("actual: %v, expected: %v", actual, expected)
	}
}
