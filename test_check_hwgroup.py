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
