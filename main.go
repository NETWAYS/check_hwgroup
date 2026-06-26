package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/NETWAYS/check_hwgroup/internal/hwgroup"

	"github.com/NETWAYS/go-check"
	"github.com/alecthomas/kong"
)

var (
	// These get filled at build time with the proper values.
	version = "development"
	commit  = "HEAD"
	date    = "latest"
)

// V3Flags are the flags for the SNMPv3
type V3Flags struct {
	Username     string `kong:"help='SNMPv3 username',default='public'"`
	AuthProto    string `kong:"help='SNMPv3 authentication protocol (MD5|SHA|SHA224|SHA256|SHA384|SHA512)',default='SHA',enum='MD5,SHA,SHA224,SHA256,SHA384,SHA512'"`
	AuthPassword string `kong:"help='SNMPv3 authentication password',env='CHECK_HWGROUP_V3_AUTH_PASS'"`
	PrivProto    string `kong:"help='SNMPv3 privacy protocol (DES|AES|AES192|AES256)',default='AES',enum='DES,AES,AES192,AES256'"`
	PrivPassword string `kong:"help='SNMPv3 privacy password',env='CHECK_HWGROUP_V3_PRIV_PASS'"`
}

type CLI struct {
	Host        string        `kong:"required,env='CHECK_HWGROUP_HOST',help='Hostname or IP of the device'"`
	Community   string        `kong:"required,default='public',env='CHECK_HWGROUP_COMMUNITY',help='SNMP community string'"`
	Port        uint16        `kong:"required,default=161,env='CHECK_HWGROUP_PORT',help='Port of the device'"`
	SNMPVersion string        `kong:"required,default='3',enum='1,2c,3',help='SNMP Version to use'"`
	Warning     string        `kong:"required,help='Warning threshold for return value'"`
	Critical    string        `kong:"required,help='Critical threshold for return value'"`
	Timeout     time.Duration `kong:"default='30s',help='Timeout for the connection'"`
	V3          V3Flags       `kong:"embed='',prefix='v3-',group='SNMPv3'"`

	Sensor  *uint `kong:"required,xor=checkType,help='Check the given sensor ID'"`
	Output  *uint `kong:"required,xor=checkType,help='Check the given relay output ID'"`
	Contact *uint `kong:"required,xor=checkType,help='check the given dry contact ID'"`

	Version kong.VersionFlag `kong:"help='Print version information and quit'"`
}

func main() {
	var cli CLI

	// Create and parse CLI flags
	kong.Parse(&cli,
		kong.Name("check_hwgroup"),
		kong.Description(`Monitoring check plugin for the HW group environmental devices`),
		kong.Vars{"version": buildVersion()},
	)

	// Parse the threshold strings
	crit, errParseCrit := check.ParseThreshold(cli.Critical)

	if errParseCrit != nil {
		check.ExitError(errParseCrit)
	}

	warn, errParseWarn := check.ParseThreshold(cli.Warning)

	if errParseWarn != nil {
		check.ExitError(errParseWarn)
	}

	// Which SNMP version to use
	SNMPVersion, errVersion := hwgroup.MapSNMPVersion(cli.SNMPVersion)

	if errVersion != nil {
		check.ExitError(errVersion)
	}

	snmpConfig := hwgroup.SNMPConfig{
		Version:      SNMPVersion,
		Community:    cli.Community,
		Username:     cli.V3.Username,
		AuthPassword: cli.V3.AuthPassword,
		AuthProto:    cli.V3.AuthProto,
		PrivPassword: cli.V3.PrivPassword,
		PrivProto:    cli.V3.PrivProto,
	}

	var client *hwgroup.Client

	var errClient error

	switch cli.SNMPVersion {
	case "1":
		client, errClient = hwgroup.NewSNMPv1Client(cli.Host, cli.Port, cli.Timeout, snmpConfig)
		if errClient != nil {
			check.ExitError(errClient)
		}
	case "2c":
		client, errClient = hwgroup.NewSNMPv2Client(cli.Host, cli.Port, cli.Timeout, snmpConfig)
		if errClient != nil {
			check.ExitError(errClient)
		}
	case "3":
		client, errClient = hwgroup.NewSNMPv3Client(cli.Host, cli.Port, cli.Timeout, snmpConfig)
		if errClient != nil {
			check.ExitError(errClient)
		}
	default:
		check.ExitError(errors.New("invalid SNMP version string"))
	}

	// Check if we support the requested device
	deviceName, errDeviceName := client.QueryDeviceName()

	if errDeviceName != nil {
		check.ExitError(errDeviceName)
	}

	deviceType, isSupported := hwgroup.IsSupportedDevice(deviceName)

	if !isSupported {
		check.ExitError(fmt.Errorf("device '%s' is not supported", deviceName))
	}

	// The result from the SNMP query
	var result hwgroup.SensorResult

	// Final output string
	var output string

	var errQuery error

	// Check the sensor if flag was set
	if cli.Sensor != nil {
		result, errQuery = client.QuerySensor(deviceType, *cli.Sensor)

		if errQuery != nil {
			check.ExitError(errQuery)
		}

		output = fmt.Sprintf("Sensor value: %.1f", result.Value)
	}

	// Check the output if flag was set
	if cli.Output != nil {
		result, errQuery = client.QueryOutput(deviceType, *cli.Output)

		if errQuery != nil {
			check.ExitError(errQuery)
		}

		if result.ContactOutputFields.Name != "" {
			output = result.ContactOutputFields.OutputString()
		}
	}

	// Check the contact if flag was set
	if cli.Contact != nil {
		result, errQuery = client.QueryContact(deviceType, *cli.Contact)

		if errQuery != nil {
			check.ExitError(errQuery)
		}

		if result.ContactOutputFields.Name != "" {
			output = result.ContactOutputFields.ContactString()
		}
	}

	// Build the final result
	var perfdata check.PerfdataList

	perfdata.Add(&check.Perfdata{
		Label: result.Name,
		Value: result.Value,
		Warn:  warn,
		Crit:  crit,
	})

	state := check.OK

	if crit.DoesViolate(result.Value) {
		state = check.Critical
	} else if warn.DoesViolate(result.Value) {
		state = check.Warning
	}

	check.ExitWithPerfdata(state, perfdata, deviceName, "-", output)
}

// buildVersion returns the current version of this tool
func buildVersion() string {
	result := version

	if commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, commit)
	}

	if date != "" {
		result = fmt.Sprintf("%s\ndate: %s", result, date)
	}

	return result
}
