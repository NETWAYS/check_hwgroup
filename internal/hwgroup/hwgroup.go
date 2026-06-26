package hwgroup

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

const (
	hwgroupBaseOID = ".1.3.6.1.4.1.21796"
	deviceNameOID  = ".1.3.6.1.2.1.1.1.0"
	poseidonType   = "Poseidon"
	damoclesType   = "Damocles"
	steType        = "STE1"
	ste2Type       = "STE2"
	wldType        = "WLD"
	unknownState   = "unknown"
)

var supportedDeviceTypes = []string{poseidonType, damoclesType, steType, ste2Type, wldType}

// OID prefixes for sensors per device type.
var sensorIDPrefix = map[string]string{
	poseidonType: "3.3.3.1.8",
	damoclesType: "3.4.3.1.8",
	ste2Type:     "4.9.3.1.8",
	steType:      "4.1.3.1.8",
	wldType:      "4.5.4.1.5",
}

// OID suffixes for (name, state, value) per device type.
var sensorPathSuffix = map[string][3]string{
	poseidonType: {"3.3.3.1.2", "3.3.3.1.4", "3.3.3.1.6"},
	damoclesType: {"3.4.3.1.2", "3.4.3.1.4", "3.4.3.1.6"},
	ste2Type:     {"4.9.3.1.2", "4.9.3.1.3", "4.9.3.1.5"},
	steType:      {"4.1.3.1.2", "4.1.3.1.3", "4.1.3.1.5"},
	wldType:      {"4.5.4.1.2", "4.5.4.1.3", "4.5.4.1.6"},
}

// deviceTypeNums maps device names to their numeric OID component used in contact/output paths.
var deviceTypeNums = map[string]int{
	damoclesType: 4,
	poseidonType: 3,
}

// SensorResult is used as a return for all Query method
type SensorResult struct {
	Name                string
	Value               float64
	ContactOutputFields ContactOutputFields
}

// ContactOutputFields is added to the SensorResult when Contact or Output is requested,
// so that we can add the extra information
type ContactOutputFields struct {
	Value      float64 // col 2: current state (0/1)
	Name       string  // col 3: user-defined label
	AlarmSetup int     // col 4: alarm config index (contact) / type index (output)
	AlarmState int     // col 5: alarm state index (contact) / mode index (output)
}

// ContactString returns a string representation of the state
// if the fields are for a dry contact
func (cf *ContactOutputFields) ContactString() string {
	var state string

	// From vendor docs:
	// Current sensor state 0 = normal, 1 = Alarm activated but not send
	switch cf.AlarmState {
	case 0:
		state = "normal"
	case 1:
		state = "activated"
	default:
		state = unknownState
	}

	var setup string

	// From vendor docs:
	// alarm settings for this Binary input – 1 byte
	// 0 = active if on, 1 = active if off, 2 = inactive
	switch cf.AlarmSetup {
	case 0:
		setup = "active if on"
	case 1:
		setup = "active if off"
	case 2:
		setup = "inactive"
	default:
		setup = unknownState
	}

	return fmt.Sprintf("Contact name: %s, AlarmState: %s, AlarmSetup: %s", cf.Name, state, setup)
}

// OutputString returns a string representation of the state
// if the fields are for a relay output
func (cf *ContactOutputFields) OutputString() string {
	var state string

	// From vendor docs:
	// 0: X/Y = On / Off (Relay output)
	// 1: X/Y = "On (+10V)" / "Off (-10V)" (RTS output)
	// 2: X/Y = "On (+10V)" / "Off (0V)" (DTR output)
	switch cf.AlarmState {
	case 0:
		state = "On / Off (Relay output)"
	case 1:
		state = "On (+10V) / Off (-10V) (RTS output)"
	case 2:
		state = "On (+10V) / Off (0V) (DTR output)"
	default:
		state = unknownState
	}

	var setup string

	// From vendor docs:
	// Output control mode (Manual / Local + condition)
	// 0 = Manual output control (value defined by Value tag)
	// 1 = Local output control (On if any alarm)
	// 2 = Local output control (On if value equal to Trigger)
	// 3 = Local output control (On if value higher than Trigger)
	// 4 = Local output control (On if value lower than Trigger)
	// 5 = Local output control (On if Alarm on)
	switch cf.AlarmSetup {
	case 0:
		setup = "Manual output control"
	case 1:
		setup = "On if any alarm"
	case 2:
		setup = "On if value equal to Trigger"
	case 3:
		setup = "On if value higher than Trigger"
	case 4:
		setup = "On if value lower than Trigger"
	case 5:
		setup = "On if Alarm on"
	default:
		setup = unknownState
	}

	return fmt.Sprintf("Output name: %s, Type: %s, Mode: %s", cf.Name, state, setup)
}

// Client is a small wrapper for gosnmp.GoSNMP so that we can
// add some convenience query methods to the type.
type Client struct {
	SNMPClient *gosnmp.GoSNMP
}

func NewClient(hostname string, port uint16, community string, snmpVersion gosnmp.SnmpVersion, timeout time.Duration) *Client {
	// TODO Support SNMPv3
	snmpClient := &gosnmp.GoSNMP{
		Target:    hostname,
		Port:      port,
		Community: community,
		Version:   snmpVersion,
		Timeout:   timeout,
		Retries:   3,
	}

	c := &Client{
		SNMPClient: snmpClient,
	}

	return c
}

func (c *Client) Close() {
	if c.SNMPClient != nil && c.SNMPClient.Conn != nil {
		c.SNMPClient.Conn.Close()
	}
}

// QueryDeviceName uses SNMP to fetch the device name
func (c *Client) QueryDeviceName() (string, error) {
	errConnect := c.SNMPClient.Connect()
	if errConnect != nil {
		return "", fmt.Errorf("could not connect to device: %w", errConnect)
	}

	defer c.SNMPClient.Close()

	value, err := c.getString(deviceNameOID)

	if err != nil {
		return "", fmt.Errorf("could not query device name: %w", err)
	}

	return value, nil
}

// QuerySensor uses SNMP to fetch the sensor value
func (c *Client) QuerySensor(deviceType string, sensorID uint) (SensorResult, error) {
	var result SensorResult

	_, ok := deviceTypeNums[deviceType]

	if !ok {
		return result, fmt.Errorf("sensor checks are not supported for %s", deviceType)
	}

	errConnect := c.SNMPClient.Connect()
	if errConnect != nil {
		return result, fmt.Errorf("could not connect to device: %w", errConnect)
	}

	defer c.SNMPClient.Close()

	sensorIndex := -1

	// Check if we got the requested sensor ID
	for _, slot := range []int{1, 2} {
		oid := fmt.Sprintf("%s.%s.%d", hwgroupBaseOID, sensorIDPrefix[deviceType], slot)

		id, err := c.getInt(oid)

		if err != nil {
			continue
		}

		if id == int(sensorID) {
			sensorIndex = slot
			break
		}
	}

	if sensorIndex == -1 {
		return result, fmt.Errorf("could not find sensor: %d", sensorID)
	}

	paths := sensorPathSuffix[deviceType]

	sensorName, err := c.getString(fmt.Sprintf("%s.%s.%d", hwgroupBaseOID, paths[0], sensorIndex))
	if err != nil {
		return result, err
	}

	sensorState, err := c.getInt(fmt.Sprintf("%s.%s.%d", hwgroupBaseOID, paths[1], sensorIndex))
	if err != nil {
		if sensorState == 0 {
			return result, fmt.Errorf("no sensor attached: %w", err)
		}

		return result, err
	}

	sensorValue, err := c.getFloat64(fmt.Sprintf("%s.%s.%d", hwgroupBaseOID, paths[2], sensorIndex))
	if err != nil {
		return result, err
	}

	result.Name = sensorName
	// The value is: Integer (decimal * 10) representation of the temperature (integer)
	// To get the actual float value we need to divide it
	result.Value = sensorValue / 10

	return result, nil
}

// QueryContact uses SNMP to fetch the contact value
func (c *Client) QueryContact(deviceType string, contactID uint) (SensorResult, error) {
	var result SensorResult

	devNum, ok := deviceTypeNums[deviceType]

	if !ok {
		return result, fmt.Errorf("contact checks are not supported for %s", deviceType)
	}

	errConnect := c.SNMPClient.Connect()
	if errConnect != nil {
		return result, fmt.Errorf("could not connect to device: %w", errConnect)
	}

	defer c.SNMPClient.Close()

	fields, err := c.getContactOutputFields("1.1", devNum, contactID)

	if err != nil {
		return result, err
	}

	result.Name = fields.Name
	result.ContactOutputFields = fields
	// From vendor docs:
	// Current value 0/1 (Read only)
	result.Value = fields.Value

	return result, nil
}

// QueryOutput uses SNMP to fetch the output value
func (c *Client) QueryOutput(deviceType string, outputID uint) (SensorResult, error) {
	var result SensorResult

	devNum, ok := deviceTypeNums[deviceType]

	if !ok {
		return result, fmt.Errorf("output checks are not supported for %s", deviceType)
	}

	errConnect := c.SNMPClient.Connect()
	if errConnect != nil {
		return result, fmt.Errorf("could not connect to device: %w", errConnect)
	}

	defer c.SNMPClient.Close()

	fields, err := c.getContactOutputFields("2.1", devNum, outputID)

	if err != nil {
		return result, err
	}

	result.Name = fields.Name
	result.ContactOutputFields = fields
	// From vendor docs:
	// 0/1 Current output value
	// 0 = Y ("Off" / "Off (-10V)" / "Off (0V)")
	// 1 = X ("On" / "On (+10V)" / "On (+10V)")
	result.Value = fields.Value

	return result, nil
}

// getContactOutputFields fetches OID columns 2–5 under the given table path
func (c *Client) getContactOutputFields(table string, devNum int, id uint) (ContactOutputFields, error) {
	var fields ContactOutputFields

	// These are float
	currentStateOID := fmt.Sprintf("%s.3.%d.%s.2.%d", hwgroupBaseOID, devNum, table, id)

	currentState, errC := c.getFloat64(currentStateOID)
	if errC != nil {
		return fields, errC
	}

	// These are strings
	userDefinedLabelOID := fmt.Sprintf("%s.3.%d.%s.3.%d", hwgroupBaseOID, devNum, table, id)

	userLabel, errU := c.getString(userDefinedLabelOID)
	if errU != nil {
		return fields, errU
	}

	// These are ints
	alarmConfigOID := fmt.Sprintf("%s.3.%d.%s.4.%d", hwgroupBaseOID, devNum, table, id)

	alarmConfig, errAC := c.getInt(alarmConfigOID)
	if errAC != nil {
		return fields, errAC
	}

	alarmStateOID := fmt.Sprintf("%s.3.%d.%s.5.%d", hwgroupBaseOID, devNum, table, id)

	alarmState, errAS := c.getInt(alarmStateOID)
	if errAS != nil {
		return fields, errAS
	}

	fields.Value = currentState
	fields.Name = userLabel
	fields.AlarmSetup = alarmConfig
	fields.AlarmState = alarmState

	return fields, nil
}

// getString uses SNMP GET and returns a string
func (c *Client) getString(oid string) (string, error) {
	result, err := c.SNMPClient.Get([]string{oid})

	if err != nil {
		return "", fmt.Errorf("could not fetch %s: %w", oid, err)
	}

	if result.Error != gosnmp.NoError {
		return "", fmt.Errorf("could not fetch %s: %v", oid, result.Error)
	}

	if len(result.Variables) == 0 {
		return "", fmt.Errorf("no value returned for %s", oid)
	}

	s := pduToString(result.Variables[0])

	if s == "" {
		return "", fmt.Errorf("no value returned for %s", oid)
	}

	return s, nil
}

// getString uses SNMP GET and returns an int
func (c *Client) getInt(oid string) (int, error) {
	s, errGet := c.getString(oid)

	if errGet != nil {
		return 0, errGet
	}

	i, errConv := strconv.Atoi(strings.TrimSpace(s))

	if errConv != nil {
		return 0, errConv
	}

	return i, nil
}

// getString uses SNMP GET and returns a float64
func (c *Client) getFloat64(oid string) (float64, error) {
	s, errGet := c.getString(oid)

	if errGet != nil {
		return 0, errGet
	}

	f, errConv := strconv.ParseFloat(strings.TrimSpace(s), 64)

	if errConv != nil {
		return 0, errConv
	}

	return f, nil
}

// MapSNMPVersion maps strings to gosnmp.SnmpVersion
func MapSNMPVersion(version string) (gosnmp.SnmpVersion, error) {
	switch version {
	case "1":
		return gosnmp.Version1, nil
	case "2c":
		return gosnmp.Version2c, nil
	case "3":
		return gosnmp.Version3, nil
	default:
		return gosnmp.Version1, errors.New("invalid SNMP version string")
	}
}

// IsSupportedDevice checks if the device is supported and returns the
// device's internal type that we use to build OIDs
func IsSupportedDevice(deviceName string) (string, bool) {
	for _, dev := range supportedDeviceTypes {
		if strings.Contains(deviceName, dev) {
			return dev, true
		}
	}

	return "", false
}

// pduToString is a simple helper to turn a PDU into a string
func pduToString(pdu gosnmp.SnmpPDU) string {
	switch v := pdu.Value.(type) {
	case []byte:
		return strings.TrimSpace(string(v))
	case string:
		return strings.TrimSpace(v)
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}
