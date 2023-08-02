#!/usr/bin/env python3

import unittest
import unittest.mock as mock
import sys

sys.path.append('..')


from check_hwgroup import commandline
from check_hwgroup import CheckHWGroupResource
from check_hwgroup import CheckHWGroupError


class CLITesting(unittest.TestCase):

    def test_commandline(self):
        actual = commandline(['-H', 'localhost', '-C', 'foobar', '-w', '5', '-c', '10', '-S', '216'])
        self.assertEqual(actual.host, 'localhost')
        self.assertEqual(actual.community, 'foobar')
        self.assertEqual(actual.critical, 10)

class CheckTesting(unittest.TestCase):

    def test_checkhw(self):

        def SNMPReq(x,y):
            return 'Poseidon 1337'

        CheckHWGroupResource.SNMPReq = SNMPReq
        check = CheckHWGroupResource('host', 'community', 1234, 'sensor', 'contact', 'output')

        self.assertEqual(check.deviceName, 'Poseidon 1337')
        self.assertEqual(check.deviceType, 'Poseidon')

    def test_checkhw_unsupported(self):

        def SNMPReq(x,y):
            return 'Foobar Bar 1337'
        CheckHWGroupResource.SNMPReq = SNMPReq

        with self.assertRaises(CheckHWGroupError) as hwerror:
            CheckHWGroupResource('host', 'community', 1234, 'sensor', 'contact', 'output')

    def test_probe_poseidon_sensor(self):
        m = mock.MagicMock()

        def side_effect(arg):
            values = {
                '.1.3.6.1.2.1.1.1.0': 'Poseidon 1337',
                '.1.3.6.1.4.1.21796.3.3.3.1.8.1': '666', # SensorID
                '.1.3.6.1.4.1.21796.3.3.3.1.8.2': '111', # SensState Poseidon
                '.1.3.6.1.4.1.21796.3.3.3.1.2.1': 'poseidon', # SensName
                '.1.3.6.1.4.1.21796.3.3.3.1.4.1': '222', # SensState
                '.1.3.6.1.4.1.21796.3.3.3.1.6.1': '333'  # SensValue
            }
            return values[arg]

        m.side_effect = side_effect

        # Monkey Patch request function with mock
        CheckHWGroupResource.SNMPReq = m

        check = CheckHWGroupResource('host', 'community', 1234, sensor=666, contact=None, output=None)

        actual = check._probe()
        expected = ('poseidon', 33.3)
        self.assertEqual(actual, expected)

    def test_probe_damocles_error(self):
        m = mock.MagicMock()

        def side_effect(arg):
            values = {
                '.1.3.6.1.2.1.1.1.0': 'Damocles',
                '.1.3.6.1.4.1.21796.3.4.3.1.8.1': 'nosuchsensor', # SensorID
                '.1.3.6.1.4.1.21796.3.4.3.1.8.2': '' # SensState
            }
            return values[arg]

        m.side_effect = side_effect

        # Monkey Patch request function with mock
        CheckHWGroupResource.SNMPReq = m

        check = CheckHWGroupResource('host', 'community', 1234, sensor=666, contact=None, output=None)

        with self.assertRaises(CheckHWGroupError):
            check._probe()

    def test_probe_poseidon_contact(self):
        m = mock.MagicMock()

        def side_effect(arg):
            values = {
                '.1.3.6.1.2.1.1.1.0': 'Poseidon 3',
                '.1.3.6.1.4.1.21796.3.3.1.1.2.666': '333', # inpValue
                '.1.3.6.1.4.1.21796.3.3.1.1.3.666': 'barfoo', # inpName
                '.1.3.6.1.4.1.21796.3.3.1.1.4.666': '0', # AlarmSetup
                '.1.3.6.1.4.1.21796.3.3.1.1.5.666': '0' # AlarmState
            }
            return values[arg]

        m.side_effect = side_effect

        # Monkey Patch request function with mock
        CheckHWGroupResource.SNMPReq = m

        check = CheckHWGroupResource('host', 'community', 1234, sensor=None, contact=666, output=None)

        actual = check._probe()
        expected =  ('barfoo [AlarmState: normal, AlarmSetup: inactive]', 333.0)
        self.assertEqual(actual, expected)

    def test_probe_poseidon_output(self):
        m = mock.MagicMock()

        def side_effect(arg):
            values = {
                '.1.3.6.1.2.1.1.1.0': 'Poseidon 3',
                '.1.3.6.1.4.1.21796.3.3.2.1.2.666': '333', # inpValue
                '.1.3.6.1.4.1.21796.3.3.2.1.3.666': 'barfoo', # inpName
                '.1.3.6.1.4.1.21796.3.3.2.1.4.666': '1', # AlarmSetup
                '.1.3.6.1.4.1.21796.3.3.2.1.5.666': '2' # AlarmState
            }
            return values[arg]

        m.side_effect = side_effect

        # Monkey Patch request function with mock
        CheckHWGroupResource.SNMPReq = m

        check = CheckHWGroupResource('host', 'community', 1234, sensor=None, contact=None, output=666)

        actual = check._probe()
        expected =  ('barfoo [Type: rts (-10V,+10V), Mode: autoTriggerEq]', 333.0)
        self.assertEqual(actual, expected)
