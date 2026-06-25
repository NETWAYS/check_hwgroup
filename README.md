# check_hwgroup

Checks the hwgroup environmental devices. Supported devices are:

* [Poseidon](https://www.hw-group.com/devices/poseidon)
* [Damocles](https://www.hw-group.com/devices/monitoring)
* [HWg-STE](https://www.hw-group.com/device/hwg-ste)

## Usage

```
Flags:
--help                  Show context-sensitive help.
--host=STRING           Hostname or IP of the device ($CHECK_HWGROUP_HOST)
--community="public"    SNMP community string ($CHECK_HWGROUP_COMMUNITY)
--port=UINT-16          Port of the device ($CHECK_HWGROUP_PORT)
--snmp-version="2c"     SNMP Version to use
--warning=STRING        Warning threshold for return value
--critical=STRING       Critical threshold for return value
--timeout=30s           Timeout for the connection
--sensor=SENSOR         Check the given sensor ID
--output=OUTPUT         Check the given relay output ID
--contact=CONTACT       check the given dry contact ID
--version               Print version information and quit
```
