#!/usr/bin/env python

# check_hwgroup.py - checks the hwgroup environmental devices
# Copyright (C) 2014  NETWAYS GmbH <http://www.netways.de>
#
# Version: 0.9
# Author: Alexander Klimov <alexander.klimov@netways.de>
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

from __future__ import print_function
import sys
from argparse import ArgumentParser
from pysnmp.entity.rfc3413.oneliner import cmdgen
import nagiosplugin

enterprise = '.1.3.6.1.4.1.21796'


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
    def __init__(
        self,
        host, community,
        sensor, contact, output,
        **kwargs
    ):
        """Prepare the resource

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
        (
            self.host,
            self.community,
            self.sensor,
            self.contact,
            self.output
        ) = (
            host,
            community,
            sensor,
            contact,
            output
        )

        self.deviceName = str(self.SNMPReq('.1.3.6.1.2.1.1.1.0'))
        for device in ('Poseidon', 'Damocles', 'STE', 'WLD'):
            if device in self.deviceName:
                self.deviceType = device
                break
        else:
            raise CheckHWGroupError(
                "Device '{}' not supported".format(self.deviceName)
            )

    def SNMPReq(self, MIBInit):
        """Send a SNMP request and return the response

        :param MIBInit: MIB (initializer)
        :type MIBInit: str
        :rtype: str
        :raise CheckHWGroupError: if a SNMP error occurs
        """
        (
            errorIndication,
            errorStatus,
            errorIndex,
            varBinds
        ) = cmdgen.CommandGenerator().getCmd(
            cmdgen.CommunityData(self.community, mpModel=0),
            cmdgen.UdpTransportTarget((self.host, 161)),
            cmdgen.MibVariable(MIBInit),
            lookupNames=True, lookupValues=True
        )
        if errorIndication:
            raise CheckHWGroupError('SNMP error: {}'.format(errorIndication))
        elif errorStatus:
            raise CheckHWGroupError('SNMP error: {}'.format(errorStatus))
        else:
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

    def _probe(self):
        if self.sensor is not None:
            for i in (1, 2):
                try:
                    sensorID = int(self.SNMPReq('{}.{}.{}'.format(
                        enterprise,
                        {
                            'Poseidon': '3.3.3.1.8',
                            'Damocles': '3.4.3.1.8',
                            'STE':      '4.1.3.1.8',
                            'WLD':      '4.5.4.1.5'
                        }[self.deviceType],
                        i
                    )))
                except (ValueError, CheckHWGroupError):
                    continue
                if sensorID == self.sensor:
                    sensorID = i
                    break
            else:
                raise CheckHWGroupError(
                    'Sensor ID ({}) not found'.format(self.sensor)
                )

            (
                sensName,
                sensState,
                sensValue
            ) = [self.SNMPReq(
                '{}.{}.{}'.format(enterprise, OID, sensorID)
            ) for OID in {
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
            }[self.deviceType]]

            if not int(sensState):
                raise CheckHWGroupError('getting sensor values failed')

            return (
                sensName,
                float(sensValue) / 10
            )
        elif self.contact is not None:
            (
                inpValue,
                inpName,
                inpAlarmSetup,
                inpAlarmState
            ) = [self.SNMPReq(
                '{}.3.4.1.1.{}.{}'.format(enterprise, OID, self.contact)
            ) for OID in xrange(2, 6)]

            return (
                '{} [AlarmState: {}, AlarmSetup: {}]'.format(
                    inpName,
                    (
                        'normal',
                        'alarm'
                    )[int(inpAlarmState)],
                    (
                        'inactive',
                        'activeOff',
                        'activeOn'
                    )[int(inpAlarmSetup)]
                ),
                float(inpValue)
            )
        else:
            (
                outValue,
                outName,
                outType,
                outMode
            ) = [self.SNMPReq(
                '{}.3.4.2.1.{}.{}'.format(enterprise, OID, self.output)
            ) for OID in xrange(2, 6)]

            return (
                '{} [Type: {}, Mode: {}]'.format(
                    outName,
                    (
                        'relay (off, on)',
                        'rts (-10V,+10V)',
                        'dtr (0V,+10V)'
                    )[int(outType)],
                    (
                        'manual',
                        'autoAlarm',
                        'autoTriggerEq',
                        'autoTriggerHi',
                        'autoTriggerLo'
                    )[int(outMode)]
                ),
                float(outValue)
            )


def main():
    argp = ArgumentParser(
        description='checks the hwgroup environmental devices'
    )
    argp.add_argument('-V', '--version', action='version', version='0.9')
    argp.add_argument('-v', '--verbose', action='count', default=0)
    for (short, long, type, default, help) in (
        ('H', 'host',      str,   object,   'The hostname or ipaddress '
                                            'of the hwgroup device'),
        ('C', 'community', str,   'public', 'The SNMP community '
                                            'of the hwgroup device'),
        ('w', 'warning',   float, object,   'Warning threshold'),
        ('c', 'critical',  float, object,   'Critical threshold')
    ):
        argp.add_argument(
            '-{}'.format(short),
            '--{}'.format(long),
            type=type,
            default=default,
            help=help,
            required=default is object
        )
    SIO = argp.add_mutually_exclusive_group(required=True)
    for (short, long, help) in (
        ('S', 'sensor',  'The sensor to check'),
        ('I', 'contact', 'The dry contact to check'),
        ('O', 'output',  'The relay output to check')
    ):
        SIO.add_argument(
            '-{}'.format(short),
            '--{}'.format(long),
            type=int,
            help=help,
            default=None
        )
    try:
        args = vars(argp.parse_args())
    except SystemExit:
        sys.exit(3)

    try:
        resource = CheckHWGroupResource(**args)
        check = nagiosplugin.Check(
            resource,
            nagiosplugin.ScalarContext(
                'check_hwgroup',
                args['warning'],
                args['critical']
            ),
        )
        check.name = resource.deviceName
        check.main(min(args['verbose'], 3))
    except CheckHWGroupError as e:
        print('ERROR:', e, file=sys.stderr)
        sys.exit(3)

if __name__ == '__main__':
    main()
