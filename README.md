# check_hwgroup

Checks the hwgroup environmental devices. Supported devices are:

* [Poseidon](https://www.hw-group.com/devices/poseidon)
* [Damocles](https://www.hw-group.com/devices/monitoring)
* [HWg-STE](https://www.hw-group.com/device/hwg-ste)

## Usage

```
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

The warning and critical flags support thresholds in the common Nagios format (e.g. `~:10`).

## Examples

### Sensors

```bash
check_hwgroup --community public --snmp-version 2c --port 1161 --host poseidon2-3266.internal --warning 10 --critical 20 --sensor 29448

[CRITICAL] - Poseidon2 3266 SNMP Supervisor v3.8.4 - Sensor value: 25.3|'HTemp Rack 19'=25.3;10;20

```

```bash
check_hwgroup --community public --snmp-version 1 --port 161 --host ste2.internal --warning 10 --critical 20 --sensor 52290

[WARNING] - STE2 r2, fw:1.5.9_2478 - Sensor value: 19.3|'Temp Rack'=19.3;10;20
```

Check for a negative value:

```bash
check_hwgroup --community public --snmp-version 2c --port 1161 --host poseidon2-3265.internal --warning="-5:" --critical="-10:" --sensor 33640

[CRITICAL] - Poseidon 3265 SNMP Supervisor v3.0.4 - Sensor value: -14.5|MORAL3=-14.5;-5:;-10:
```

### Dry Contacts

```bash
check_hwgroup --community public --snmp-version 2c --port 1161 --host poseidon2-3266.internal --warning 1 --critical 1 --contact 1

[OK] - Poseidon2 3266 SNMP Supervisor v3.8.4 - Contact name: Binary 1, SensorState: 0, AlarmSetup: active if on|'Binary 1'=0;1;1
```

```bash
check_hwgroup --community public --snmp-version 1 --port 161 --host damocles-mini.internal --warning 1 --critical 1 --contact 1

[OK] - Damocles MINI SNMP Supervisor v1.0.11 - Contact name: Input 1, SensorState: 0, AlarmSetup: inactive|'Input 1'=0;1;1
```

How the plugin maps the returned contact states:

- `0`: 0
- `1`: 1
- `everything else`: unknown

How the plugin maps the returned contact setup:

- `0`: active if on
- `1`: active if off
- `2`: inactive
- `everything else`: unknown

### Relay Outputs

```bash
check_hwgroup --community public --snmp-version 2c --port 1161 --host poseidon2-3266.internal --warning 1 --critical 1 --output 1

[OK] - Poseidon2 3266 SNMP Supervisor v3.8.4 - Output name: VirtBinOut 1, Type: On / Off (Relay output), Mode: Manual output control|'VirtBinOut 1'=0;1;1
```

How the plugin maps the returned output states:

- `0`: On / Off (Relay output)
- `1`: On (+10V) / Off (-10V) (RTS output)
- `2`: On (+10V) / Off (0V) (DTR output)
- `everything else`: unknown

How the plugin maps the returned output setup:

- `0`: Manual output control
- `1`: On if any alarm
- `2`: On if value equal to Trigger
- `3`: On if value higher than Trigger
- `4`: On if value lower than Trigger
- `5`: On if Alarm on
- `everything else`: unknown