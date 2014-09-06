#!/usr/bin/perl -w

# ------------------------------------------------------------------------------
# check_hwgroup.pl - checks the hwgroup environmental devices.
# Copyright (C) 2009-2012  NETWAYS GmbH, www.netways.de
# 
# Author: Thomas Gelf <thomas.gelf@netways.de>
# Author: Michael Streb <michael.streb@netways.de>
# Author: Bernd Löhlein <bernd.loehlein@netways.de>
# Author: Marius Hein <marius.hein@netways.de>
#
# This program is free software; you can redistribute it and/or
# modify it under the tepdu of the GNU General Public License
# as published by the Free Software Foundation; either version 2
# of the License, or (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program; if not, write to the Free Software
# Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
# ------------------------------------------------------------------------------

# basic requirements
use strict;
use Getopt::Long;
use File::Basename;
use Pod::Usage;
use Net::SNMP;

# predeclared subs
use subs qw(
    print_help
    check_threshold
    threshold_state
    threshold_label
);

# predeclared vars
use vars qw (
  $PROGNAME
  $VERSION

  %states
  %state_names

  $opt_host
  $opt_community
  $opt_sensor
  $opt_contact
  $opt_output
  $opt_warning
  $opt_critical

  $opt_help
  $opt_man
  $opt_version

  $module

  $device
  $response
  $output
  @oids
);

# Main values
$PROGNAME = basename($0);
$VERSION  = '1.2';

# Nagios exit states
%states = (
    OK       => 0,
    WARNING  => 1,
    CRITICAL => 2,
    UNKNOWN  => 3
);

# Nagios state names
%state_names = (
    0 => 'OK',
    1 => 'WARNING',
    2 => 'CRITICAL',
    3 => 'UNKNOWN'
);

# SNMP
my $opt_community = "public";
my $snmp_version  = "1";

#my $response;

# Get the options from cl
Getopt::Long::Configure('bundling');
GetOptions(
    'h'       => \$opt_help,
    'help'    => \$opt_help,
    'H=s'     => \$opt_host,
    'C=s',    => \$opt_community,
    'S=n',    => \$opt_sensor,
    'I=n',    => \$opt_contact,
    'O=n',    => \$opt_output,
    'w=s'     => \$opt_warning,
    'c=s'     => \$opt_critical,
    'man'     => \$opt_man,
    'V'       => \$opt_version
  )
  || print_help( 1, 'Please check your options!' );

# If somebody wants to the help ...
if ($opt_help) {
    print_help(1);
}
elsif ($opt_man) {
    print_help(2);
}
elsif ($opt_version) {
    print_help(-1);
}

# oids
my $deviceid = '.1.3.6.1.2.1.1.1.0';
my $enterprise = '.1.3.6.1.4.1.21796';

# Check if all needed options present.
unless ( $opt_host && ( $opt_sensor || $opt_contact || $opt_output ) ) {

    print_help( 1, 'Not enough options specified!' );
}
else {

    # Threshold check
    my $system_threshold = 1;
    
    if (defined($opt_warning) && defined($opt_critical)) {
        $system_threshold = 0;
    }
    
    if (defined($opt_warning) and !defined($opt_critical)) {
        print_help( 1, 'Please use both threshold switches!' );
    }
    
    if (!defined($opt_warning) and defined($opt_critical)) {
        print_help( 1, 'Please use both threshold switches!' );
    }
    
    # Open SNMP Session
    my ( $session, $error ) = Net::SNMP->session(
        -hostname  => $opt_host,
        -community => $opt_community,
        -port      => 161,
        -version   => $snmp_version
    );

    # SNMP Session failed
    if ( !defined($session) ) {
        print $state_names{ ( $states{UNKNOWN} ) } . ": $error";
        exit $states{UNKNOWN};
    }

    # request for sensor
    if (defined $opt_sensor) {

        my ($sensor_tree_id, $sensor_id);

        # Sensor states
        my %sensor_states = ();

        # Sensor OID setting

        # check for device type Poseidon/STE
        $response = $session->get_request($deviceid);
        $device = $response->{$deviceid};

        if ($device =~ m/Poseidon/i) {
            # Sensor states
            %sensor_states = (
                0 => 'invalid',
                1 => 'normal',
                2 => 'alarmstate',
                3 => 'alarm',
            );
            $sensor_tree_id = $enterprise.".3.3.3.1.8";
        } elsif ($device =~ m/Damocles/i) {
            # Sensor states
            %sensor_states = (
                0 => 'invalid',
                1 => 'normal',
                2 => 'alarmstate',
                3 => 'alarm',
            );
            $sensor_tree_id = $enterprise.".3.4.3.1.8";
            
        } elsif ($device =~ m/STE/i) {
            # Sensor states
            %sensor_states = (
                0 => 'invalid',
                1 => 'normal',
                2 => 'outofrangelo',
                3 => 'outofrangehi',
                4 => 'alarmlo',
                5 => 'alarmhi',
            );
            $sensor_tree_id = $enterprise.".4.1.3.1.8";
        } elsif ($device =~ m/WLD/i) {
            # Sensor states
            %sensor_states = (
                0 => 'invalid',
                1 => 'normal',
                3 => 'alarm'
            );
            $sensor_tree_id = $enterprise.".4.5.4.1.5";
        } else {
            print "ERROR: Device not supported\n";
            $session->close();
            exit $states{UNKNOWN};
        }

        # get the sensor ID
        $response = $session->get_table($sensor_tree_id);
        foreach my $sensor (keys %$response) {
            if($response->{$sensor} eq $opt_sensor) {
                $sensor =~ m/.?[\d\.]+(\d)$/;
                $sensor_id = $1;
            };
        }

        # check for numeric sensor id
        if (!defined $sensor_id || $sensor_id !~ m/\d/ ) {
            print "ERROR: Sensor ID not found\n";
            $session->close();
            exit $states{UNKNOWN};
        }

        # setting per device OIDs for sensor values
        if ($device =~ m/Poseidon/i) {
            push(@oids, $enterprise.'.3.3.3.1.2.'.$sensor_id);          # POSEIDON-MIB::sensName
            push(@oids, $enterprise.'.3.3.3.1.4.'.$sensor_id);          # POSEIDON-MIB::sensState
            push(@oids, $enterprise.'.3.3.3.1.5.'.$sensor_id);          # POSEIDON-MIB::sensString
            push(@oids, $enterprise.'.3.3.3.1.6.'.$sensor_id);          # POSEIDON-MIB::sensValue
            push(@oids, $enterprise.'.3.3.3.1.9.'.$sensor_id);          # POSEIDON-MIB::sensUnit
            push(@oids, $enterprise.'.3.3.99.1.2.1.6.'.$sensor_id);     # POSEIDON-MIB::sensLimitMin
            push(@oids, $enterprise.'.3.3.99.1.2.1.7.'.$sensor_id);     # POSEIDON-MIB::sensLimitMax
        } elsif ($device =~ m/Damocles/i) {
            push(@oids, $enterprise.'.3.4.3.1.2.'.$sensor_id);          # POSEIDON-MIB::sensName
            push(@oids, $enterprise.'.3.4.3.1.4.'.$sensor_id);          # POSEIDON-MIB::sensState
            push(@oids, $enterprise.'.3.4.3.1.5.'.$sensor_id);          # POSEIDON-MIB::sensString
            push(@oids, $enterprise.'.3.4.3.1.6.'.$sensor_id);          # POSEIDON-MIB::sensValue
            push(@oids, $enterprise.'.3.4.3.1.9.'.$sensor_id);          # POSEIDON-MIB::sensUnit
            push(@oids, $enterprise.'.3.4.99.1.2.1.6.'.$sensor_id);     # POSEIDON-MIB::sensLimitMin
            push(@oids, $enterprise.'.3.4.99.1.2.1.7.'.$sensor_id);     # POSEIDON-MIB::sensLimitMax
        } elsif ($device =~ m/STE/i) {
            push(@oids, $enterprise.'.4.1.3.1.2.'.$sensor_id);          # POSEIDON-MIB::sensName
            push(@oids, $enterprise.'.4.1.3.1.3.'.$sensor_id);          # POSEIDON-MIB::sensState
            push(@oids, $enterprise.'.4.1.3.1.4.'.$sensor_id);          # POSEIDON-MIB::sensString
            push(@oids, $enterprise.'.4.1.3.1.5.'.$sensor_id);          # POSEIDON-MIB::sensValue
        } elsif ($device =~ m/WLD/i) {
            push(@oids, $enterprise.'.4.5.4.1.2.'.$sensor_id);          # POSEIDON-MIB::sensName
            push(@oids, $enterprise.'.4.5.4.1.3.'.$sensor_id);          # POSEIDON-MIB::sensState
            push(@oids, $enterprise.'.4.5.4.1.5.'.$sensor_id);          # whatever
            push(@oids, $enterprise.'.4.5.4.1.6.'.$sensor_id);          # POSEIDON-MIB::sensValue
        } else {
            print "ERROR: device not supported";
            $session->close();
            exit $states{UNKNOWN};
        }

        # getting the sensor values from the device
        $response = $session->get_request(-varbindlist => \@oids) 
            or die "ERROR while getting Sensor values";

        if ($device =~ m/WLD/i) {
            if (defined $response->{$oids[3]}) {
                my @wldvals = ('normal', 'flooded', 'disconnect', 'invalid');
                $response->{$oids[2]} = $wldvals[$response->{$oids[3]}];
            }
        }
        # setting the output string
        $output .= "Sensor: ".$response->{$oids[0]}.", ";

        if ($system_threshold eq 1)  {
            $output .= "State: ".$sensor_states{$response->{$oids[1]}}.", ";
        } else {
            $output .= "State: ". threshold_label($response->{$oids[2]}). ", ";
        }
        
        $output .= "Value: ".$response->{$oids[2]};

        $output .= "| '$response->{$oids[0]}'=" . ($response->{$oids[3]} / 10) . ";";

        # append thresholds to perfdata if device is Poseidon
        if ($device =~ m/Poseidon/i) {
            $output .= ($response->{$oids[5]} / 10) . ";";
            $output .= ($response->{$oids[6]} / 10) . ";";
        }
    }

    # request for dry contact
    if (defined $opt_contact) {

        # Input states
        my %alarm_states = (
            0 => 'normal',
            1 => 'alarm',
        );
        my %input_values = (
            0 => 'off',
            1 => 'on',
        );
        my %alarm_setup = (
            0 => 'inactive',
            1 => 'activeOff',
            2 => 'activeOn',
        );

        # get the contact values
        push(@oids, $enterprise.'.3.4.1.1.3.'.$opt_contact);            # POSEIDON-MIB::inpName
        push(@oids, $enterprise.'.3.4.1.1.2.'.$opt_contact);            # POSEIDON-MIB::inpValue
        push(@oids, $enterprise.'.3.4.1.1.4.'.$opt_contact);            # POSEIDON-MIB::inpAlarmSetup
        push(@oids, $enterprise.'.3.4.1.1.5.'.$opt_contact);            # POSEIDON-MIB::inpAlarmState

        # getting the values from the device
        $response = $session->get_request(-varbindlist => \@oids) 
                        or print "ERROR: Sensor ID not found\n";
        exit $states{UNKNOWN} if !defined $response;

        # setting the output string
        $output .= "Input: ".$response->{$oids[0]}.", ";
        $output .= "AlarmState: ".$alarm_states{$response->{$oids[3]}}.", ";
        $output .= "AlarmSetup: ".$alarm_setup{$response->{$oids[2]}}.", ";
        $output .= "Value: ".$input_values{$response->{$oids[1]}};
    }
    
    if (defined $opt_output) {
        
        # output states
        my %output_values = (
            0 => 'off',
            1 => 'on',
        );
        
        # output types
        my %output_types = (
            0 => 'relay (off, on)',
            1 => 'rts (-10V,+10V)',
            2 => 'dtr (0V,+10V)',
        );

        # output mode
        my %output_modes = (
            0 => 'manual',
            1 => 'autoAlarm',
            2 => 'autoTriggerEq',
            3 => 'autoTriggerHi',
            4 => 'autoTriggerLo',
        );

        # get the contact values
        push (@oids, $enterprise.'.3.4.2.1.2.'.$opt_output);            # POSEIDON-MIB::outValue
        push (@oids, $enterprise.'.3.4.2.1.3.'.$opt_output);            # POSEIDON-MIB::outName
        push (@oids, $enterprise.'.3.4.2.1.4.'.$opt_output);            # POSEIDON-MIB::outType
        push (@oids, $enterprise.'.3.4.2.1.5.'.$opt_output);            # POSEIDON-MIB::outMode
    
        # getting the values from the device
        $response = $session->get_request(-varbindlist => \@oids)
            or print "ERROR: Sensor ID not found\n";
        exit $states{UNKNOWN} if !defined $response;

        # setting the output string
        $output .= "Output: ".$response->{$oids[1]}.", ";
        $output .= "Type: ".$output_types{$response->{$oids[2]}}.", ";
        $output .= "Mode: ".$output_modes{$response->{$oids[3]}}.", ";
        $output .= "Value: ".$output_values{$response->{$oids[0]}};
    }

    # finally close SNMP session
    $session->close();

    # print the gathered data
    print $output."\n";

    # setting exit states
    if (defined $opt_sensor) {
        if ( $device =~ m/Poseidon/i ) {
            if ($response->{$oids[1]} == 3) {
                exit $states{CRITICAL};
            } elsif ($response->{$oids[1]} == 2) {
                exit $states{WARNING};
                } elsif ($response->{$oids[1]} == 0) {
                exit $states{UNKNOWN};
            } else {
                exit $states{OK};
            }
        } elsif ($device =~ m/STE/i ) {
            if ($system_threshold eq 1) {
                if ($response->{$oids[1]} > 1) {
                    exit $states{CRITICAL};
                } elsif ($response->{$oids[1]} == 0) {
                    exit $states{UNKNOWN};
                } else {
                    exit $states{OK};
                }
            } else {
                my $state = threshold_state($response->{$oids[2]});
                exit($state);
            }
        } elsif ($device =~ m/WLD/i ) {
            if ($response->{$oids[1]} != 1) {
                exit $states{CRITICAL};
            } else {
                exit $states{OK};
            }
        }
    }

    # check for dry contacts
    if (defined $opt_contact) {
        if ($response->{$oids[3]} == 1) {
            exit $states{CRITICAL};
        } else {
            exit $states{OK};
        }
    }
}   

# -------------------------
# THE SUBS:
# -------------------------


# print_help($level, $msg);
# prints some message and the POD DOC
sub print_help {
    my ( $level, $msg ) = @_;
    $level = 0 unless ($level);
    if($level == -1) {
        print "$PROGNAME - Version: $VERSION\n";
        exit ( $states{UNKNOWN});
    }
    pod2usage(
        {
            -noperldoc => 1,
            -message => $msg,
            -verbose => $level
        }
    );

    exit( $states{UNKNOWN} );
}

# check_threshold( $syntax, $value )
# Test values against nagios threshold syntax formats
# http://nagiosplug.sourceforge.net/developer-guidelines.html#THRESHOLDFORMAT
sub check_threshold {
    my $syntax = shift;
    my $value = shift;
    
    my @check = split(/:/, $syntax);
    
    my $inside = 0;
    
    if (scalar(@check) eq 1) {
        splice(@check, 0, 0, 0);
    }
    
    if ($check[0] eq "") {
        $check[0] = 0;
    }
    
    if ($check[1] eq "") {
        $check[1] = 0;
    }
    
    if ($check[0] =~ m/^\@/) {
        $inside = 1;
        $check[0] =~ s/^\@//g;
    }
    
    # Hack, but works simple and fast
    $check[0] =~ s/^~$/-9999999/g;
    $check[1] =~ s/^~$/9999999/g;
    
    if ($inside) {
        if ($value >= $check[0] && $value <= $check[1]) {
                return 1;
        }
    } else {
        if ($value < $check[0] || $value > $check[1]) {
                return 1;
        }
    }
    
    return 0;
}

# threshold_label( $value )
# Returns a status label matching against critical
# and warning thresholds
sub threshold_label {
  my $value = shift;
  my $state = threshold_state($value);
  return $state_names{$state};
}

# threshold_state( $value )
# Returns a icinga/nagios exit state matching
# the critical and warnin thresholds
sub threshold_state {
  my $value = shift;
  
  if (check_threshold($opt_critical, $value)) {
      return 2;
  }
  elsif (check_threshold($opt_warning, $value)) {
      return 1;
  }
  
  return 0;
  
}

1;

__END__

=head1 NAME

check_hwgroup.pl - Checks hwgroup devices

=head1 SYNOPSIS

check_hwgroup.pl -h

check_hwgroup.pl -H <host> ( -S <sensor id> | -I <input id> | -O <output id> )

=head1 DESCRIPTION

Bcheck_hwgroup.pl recieves the data from the hwgroup devices. 

=head1 OPTIONS

=over 9

=item B<-h>

Display this helpmessage.

=item B<-H>

The hostname or ipaddress of the hwgroup device.

=item B<-C>

The snmp community of the hwgroup device.

=item B<-S>

The sensor to check

=item B<-I>

The dry contact to check

=item B<-O>

The relay output to check

=item B<--man>

Displays the complete perldoc manpage.

=item B<-c>

Critical threshold (only for STE device). Please see
the full man page for explanations

=item B<-w>

Warning threshold (only for STE device). Please see
the full man page for explanations

=back

=head1 SEE ALSO

Threshold formats based on the nagios plugin development guidelines:

http://nagiosplug.sourceforge.net/developer-guidelines.html#THRESHOLDFORMAT

The STE device have build in thresholds. You can override these by using the
-c and -w switch (both must be set)

=cut

