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
check_hwgroup.py -h|--help

check_hwgroup.py -V|--version

short | long        | type    | default    | description
------+-------------+---------+------------+----------------------------------
-v    | --verbose   | N/A     | N/A        | be verbose (specify up to 3 times
      |             |         |            | to increase verbosity level)
-H    | --host      | string  | (required) | the hostname or ipaddress
      |             |         |            | of the hwgroup device
-C    | --community | string  | public     | the SNMP community
      |             |         |            | of the hwgroup device
-P    | --port      | integer | 161        | the port
      |             |         |            | of the hwgroup device
-w    | --warning   | float   | (required) | warning threshold
-c    | --critical  | float   | (required) | critical threshold
-S    | --sensor    | integer | (none)     | the sensor to check
-I    | --contact   | integer | (none)     | the dry contact to check
-O    | --output    | integer | (none)     | the relay output to check
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
