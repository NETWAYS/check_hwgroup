**Note:** This is an older version and will no longer be maintained.

# check_hwgroup

Checks the hwgroup environmental devices. Supported devices are:

* [Poseidon](https://www.hw-group.com/devices/poseidon)
* [Damocles](https://www.hw-group.com/devices/monitoring)
* [HWg-STE](https://www.hw-group.com/device/hwg-ste)

# Installation

The plugin requires at least Python 3.

## Required Python Libraries

* pysnmplib
* nagiosplugin

# Usage

The options `-S`, `-I` and `-O` are mutually exclusive.

```bash
Usage: check_hwgroup --host=STRING --community="public" --port=161 --snmp-version="3" --warning=STRING --critical=STRING --sensor=SENSOR --output=OUTPUT --contact=CONTACT [flags]

Monitoring check plugin for the HW group environmental devices

Flags:
  -h, --help                  Show context-sensitive help.
  -H, --host=STRING           Hostname or IP of the device ($CHECK_HWGROUP_HOST)
  -C, --community="public"    SNMP community string ($CHECK_HWGROUP_COMMUNITY)
  -p, --port=161              Port of the device ($CHECK_HWGROUP_PORT)
  -P, --snmp-version="3"      SNMP Version to use
  -w, --warning=STRING        Warning threshold for return value
  -c, --critical=STRING       Critical threshold for return value
  -t, --timeout=15s           Timeout for the connection
  -r, --retries=2             Retries for the connection
  -s, --sensor=SENSOR         Check the given sensor ID
  -o, --output=OUTPUT         Check the given relay output ID
  -d, --contact=CONTACT       Check the given dry contact ID
  -V, --version               Print version information and quit

SNMPv3
  --v3-username="public"       SNMPv3 username
  --v3-auth-proto="SHA"        SNMPv3 authentication protocol (MD5|SHA|SHA224|SHA256|SHA384|SHA512)
  --v3-auth-password=STRING    SNMPv3 authentication password ($CHECK_HWGROUP_V3_AUTH_PASS)
  --v3-priv-proto="AES"        SNMPv3 privacy protocol (DES|AES|AES192|AES256)
  --v3-priv-password=STRING    SNMPv3 privacy password ($CHECK_HWGROUP_V3_PRIV_PASS)
```

# Examples

Fetch a sensor value and check whether it's OK, WARNING or CRITICAL:

    $ ./check_hwgroup.py -H 192.168.144.120 -S 17946 -w 30 -c 35
    POSEIDON 2250 SNMP SUPERVISOR V1.0.13 OK - Sensor 240 is 25.8 | 'Sensor 240'=25.8;30.0;35.0
    $ echo $?
    0

    $ ./check_hwgroup.py -H 192.168.144.120 -S 17946 -w 25.7 -c 35
    POSEIDON 2250 SNMP SUPERVISOR V1.0.13 WARNING - Sensor 240 is 25.8 (outside range 0:25.7) | 'Sensor 240'=25.8;25.7;35.0
    $ echo $?
    1

    $ ./check_hwgroup.py -H 192.168.144.120 -S 17946 -w 25.7 -c 25.7
    POSEIDON 2250 SNMP SUPERVISOR V1.0.13 CRITICAL - Sensor 240 is 25.8 (outside range 0:25.7) | 'Sensor 240'=25.8;25.7;25.7
    $ echo $?
    2

## Frequently occurred user errors

Wrong host/port/community:

    $ ./check_hwgroup.py -H 192.168.144.121 -S 17946 -w 30 -c 35 -P 162 -C apple
    ERROR: SNMP error: No SNMP response received before timeout
    $ echo $?
    3

Wrong sensor:

    $ ./check_hwgroup.py -H 192.168.144.120 -S 17947 -w 30 -c 35
    ERROR: Sensor ID (17947) not found
    $ echo $?
    3
