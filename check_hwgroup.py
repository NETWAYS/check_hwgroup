#!/usr/bin/env python

# check_hwgroup.py - checks the hwgroup environmental devices
# Copyright (C) 2014  NETWAYS GmbH <http://www.netways.de>
#
# Version: 1.0
# Author: Alexander Klimov <alexander.klimov@netways.de>
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 2 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.


import sys
from argparse import ArgumentParser
from pysnmp.entity.rfc3413.oneliner import cmdgen
import nagiosplugin

__version__ = '1.2.0'

enterprise = '.1.3.6.1.4.1.21796'

sensor_ids = {
    'Poseidon': '3.3.3.1.8',
    'Damocles': '3.4.3.1.8',
    'STE2':     '4.9.3.1.8',
    'STE':      '4.1.3.1.8',
    'WLD':      '4.5.4.1.5'
}

sensor_paths = {
    'Poseidon': (
        '3.3.3.1.2',
        '3.3.3.1.4',
        '3.3.3.1.6'
    ),
    'Damocles': (
        '3.4.3.1.2',
        '3.4.3.1.4',
        '3.4.3.1.6'
    ),
    'STE2': (
        '4.9.3.1.2',
        '4.9.3.1.3',
        '4.9.3.1.5'
    ),
    'STE': (
        '4.1.3.1.2',
        '4.1.3.1.3',
        '4.1.3.1.5'
    ),
    'WLD': (
        '4.5.4.1.2',
        '4.5.4.1.3',
        '4.5.4.1.6'
    )
}

class CheckHWGroupError(Exception):
    """
    CheckHWGroupError MAY be raised at runtime
    when a check_hwgroup-specific error occurs.
    CheckHWGroupError SHOULD be caught by an except statement.
    """


class CheckHWGroupResource(nagiosplugin.Resource):
    """
    The resource for fetching sensor/contact/output values
    """
    def __init__(self, host, community, port, sensor, contact, output):
        """
        Prepare the resource
        :param host: target host
        :type host: str
        :param community: SNMP community
        :type community: str
        :param sensor: target sensor
        :type sensor: int
        :param contact: target contact
        :type contact: int
        :param output: target output
        :type output: int
        """
        self.host = host
        self.community = community
        self.port = port
        self.sensor = sensor
        self.contact = contact
        self.output = output

        self.deviceName = str(self.SNMPReq('.1.3.6.1.2.1.1.1.0'))
        self.notSupported = CheckHWGroupError("Device '{}' not supported".format(self.deviceName))

        supportedDevices = ('Poseidon', 'Damocles', 'STE2', 'STE', 'WLD')
        # Check if supported Device string is in deviceName, if empty list throw error
        for device in supportedDevices:
            if device in self.deviceName:
                self.deviceType = device
                break
        else:
            raise self.notSupported

    def SNMPReq(self, MIBInit):
        """Send a SNMP request and return the response

        :param MIBInit: MIB (initializer)
        :type MIBInit: str
        :rtype: str
        :raise CheckHWGroupError: if a SNMP error occurs
        """
        errorIndication, errorStatus, _, varBinds = cmdgen.CommandGenerator().getCmd(
            cmdgen.CommunityData(self.community, mpModel=0),
            cmdgen.UdpTransportTarget((self.host, self.port)),
            cmdgen.MibVariable(MIBInit),
            lookupNames=True,
            lookupValues=True)

        if errorIndication:
            raise CheckHWGroupError('SNMP error: {}'.format(errorIndication))

        if errorStatus:
            raise CheckHWGroupError('SNMP error: {}'.format(errorStatus))

        return str(varBinds[0][1])

    def probe(self):
        """Fetch values

        :rtype: nagiosplugin.Metric
        :raise CheckHWGroupError: various reasons, see error messages below
        """
        yield nagiosplugin.Metric(
            *self._probe(),
            context='check_hwgroup'
        )

    def _probe_sensor(self):
        for ident in (1, 2):
            try:
                sensorID = int(self.SNMPReq('{}.{}.{}'.format(enterprise, sensor_ids[self.deviceType], ident)))
            except (ValueError, CheckHWGroupError):
                continue
            if sensorID == self.sensor:
                sensorID = ident
                break
        else:
            raise CheckHWGroupError('Sensor ID ({}) not found'.format(self.sensor))

        (sensName, sensState, sensValue) = [self.SNMPReq('{}.{}.{}'.format(enterprise, OID, sensorID)) for OID in sensor_paths[self.deviceType]]

        if not int(sensState):
            raise CheckHWGroupError('getting sensor values failed')

        return (sensName, float(sensValue) / 10)

    def _probe_contact(self):
        (inpValue, inpName, inpAlarmSetup, inpAlarmState) = [self.SNMPReq('{}.3.{}.1.1.{}.{}'.format(enterprise,{'Damocles': 4, 'Poseidon': 3}[self.deviceType], OID, self.contact)) for OID in range(2, 6)]

        contactName = '{} [AlarmState: {}, AlarmSetup: {}]'.format(
            inpName,
            ('normal', 'alarm')[int(inpAlarmState)],
            ('inactive', 'activeOff', 'activeOn')[int(inpAlarmSetup)])

        return (contactName, float(inpValue))

    def _probe_output(self):
        (outValue, outName, outType, outMode) = [self.SNMPReq('{}.3.{}.2.1.{}.{}'.format(enterprise, {'Damocles': 4, 'Poseidon': 3}[self.deviceType], OID, self.output)) for OID in range(2, 6)]

        outputName = '{} [Type: {}, Mode: {}]'.format(
            outName,
            ('relay (off, on)', 'rts (-10V,+10V)', 'dtr (0V,+10V)')[int(outType)],
            ('manual', 'autoAlarm','autoTriggerEq', 'autoTriggerHi', 'autoTriggerLo')[int(outMode)])

        return (outputName, float(outValue))

    def _probe(self):
        if self.sensor:
            return self._probe_sensor()
        if not self.sensor and self.contact:
            return self._probe_contact()
        if not self.sensor and not self.contact:
            return self._probe_output()

def commandline(args):
    argp = ArgumentParser(description='Check Plugin for the hwgroup environmental devices.')

    argp.add_argument('-V', '--version', action='version', version=__version__)
    argp.add_argument('-v', '--verbose', action='count', default=0)
    argp.add_argument('-H', '--host', type=str, required=True,
                      help='The hostname or ipaddress of the hwgroup device')
    argp.add_argument('-C', '--community', type=str, default='public',
                      help='The SNMP community of the hwgroup device')
    argp.add_argument('-P', '--port', default=161,
                      help='The port of the hwgroup device')
    argp.add_argument('-w', '--warning', type=float, required=True,
                      help='Warning threshold')
    argp.add_argument('-c', '--critical', type=float, required=True,
                      help='Critical threshold')

    SIO = argp.add_mutually_exclusive_group(required=True)
    SIO.add_argument('-S', '--sensor', type=int, help='The sensor to check')
    SIO.add_argument('-I', '--contact', type=int, help='The dry contact to checkThe sensor to')
    SIO.add_argument('-O', '--output', type=int, help='The relay output to check')

    return argp.parse_args(args)


def main(args):
    try:
        resource = CheckHWGroupResource(args.host, args.community, args.port, args.sensor, args.contact, args.output)
        check = nagiosplugin.Check(
            resource,
            nagiosplugin.ScalarContext('check_hwgroup', args.warning, args.critical),
        )
        check.name = resource.deviceName
        check.main(min(args.verbose, 3))
    except CheckHWGroupError as e:
        print("[UNKNOWN] - Error: %s" % (str(e)))
        return 3


if __name__ == '__main__': # pragma: no cover
    try:
        ARGS = commandline(sys.argv[1:])
        sys.exit(main(ARGS))
    except SystemExit:
        # Re-throw the exception
        raise sys.exc_info()[1].with_traceback(sys.exc_info()[2]) # pylint: disable=raise-missing-from
    except: # pylint: disable=bare-except
        print("[UNKNOWN] - Error: %s" % (str(sys.exc_info()[1])))
        sys.exit(3)
