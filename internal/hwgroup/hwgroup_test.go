package hwgroup

import (
	"strings"
	"testing"
	"time"

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
	cf := ContactOutputFields{Name: "Unit", SensorState: 0, AlarmSetup: 0}

	expected := "Contact name: Unit, SensorState: 0, AlarmSetup: active if on"

	if actual := cf.ContactString(); actual != expected {
		t.Errorf("actual: %v, expected: %v", actual, expected)
	}
}

func TestContactOutputFields_OutputString(t *testing.T) {
	cf := ContactOutputFields{Name: "Unit", SensorState: 2, AlarmSetup: 3}

	expected := "Output name: Unit, Type: On (+10V) / Off (0V) (DTR output), Mode: On if value higher than Trigger"

	if actual := cf.OutputString(); actual != expected {
		t.Errorf("actual: %v, expected: %v", actual, expected)
	}
}

func TestNewSNMPv1Client(t *testing.T) {
	config := SNMPConfig{Community: "u"}

	actual, err := NewSNMPv1Client("127.0.0.1", 161, time.Second, 2, config)

	if err != nil {
		t.Errorf("got error, expected nil: %v", err)
	}

	if actual.SNMPClient.Version != gosnmp.Version1 {
		t.Errorf("expected SNMP v1, got %v", actual.SNMPClient.Version)
	}
}

func TestNewSNMPv2Client(t *testing.T) {
	config := SNMPConfig{Community: "u"}

	actual, err := NewSNMPv2Client("127.0.0.1", 161, time.Second, 2, config)

	if err != nil {
		t.Errorf("got error, expected nil: %v", err)
	}

	if actual.SNMPClient.Version != gosnmp.Version2c {
		t.Errorf("expected SNMP v2c, got %v", actual.SNMPClient.Version)
	}
}

func TestNewSNMPv3Client(t *testing.T) {
	tests := []struct {
		name        string
		input       SNMPConfig
		expectAnErr bool
		expectedErr string
	}{
		{"v3 noauth", SNMPConfig{Version: gosnmp.Version3, Username: "u"}, false, ""},
		{"v3 auth", SNMPConfig{Version: gosnmp.Version3, Username: "u", AuthPassword: "p", AuthProto: "SHA"}, false, ""},
		{"v3 authpriv", SNMPConfig{Version: gosnmp.Version3, Username: "u", AuthPassword: "p", PrivPassword: "p", AuthProto: "MD5", PrivProto: "DES"}, false, ""},
		{"v3 no username", SNMPConfig{Version: gosnmp.Version3}, true, "username is required"},
		{"v3 priv no auth", SNMPConfig{Version: gosnmp.Version3, Username: "u", PrivPassword: "p"}, true, "auth-password is required"},
		{"v3 bad auth proto", SNMPConfig{Version: gosnmp.Version3, Username: "u", AuthPassword: "p", AuthProto: "INVALID"}, true, "unknown auth protocol: INVALID"},
		{"v3 bad priv proto", SNMPConfig{Version: gosnmp.Version3, Username: "u", AuthPassword: "p", AuthProto: "MD5", PrivPassword: "p", PrivProto: "INVALID"}, true, "unknown priv protocol: INVALID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSNMPv3Client("127.0.0.1", 161, time.Second, 2, tt.input)

			if (err != nil) != tt.expectAnErr {
				t.Errorf("actual: %v, expected: %v", err, tt.expectedErr)
				return
			}

			if tt.expectAnErr && !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("actual: %v, expected: %v", err, tt.expectedErr)
			}
		})
	}
}
