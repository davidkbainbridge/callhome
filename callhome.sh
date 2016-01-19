#!/bin/bash
# Copyright 2016 Ciena Corporation
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# you may obtain a copy of the License at
#
#    http://www.apache.org/license/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, sofware
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

PROG=`basename $0`

INTERFACE="ma1"
PORT="4321"
URI="callhome"

function usage() {
    echo "$PROG [options]"
    echo "-h, --help                display this message"
    echo "-i, --interface <name>    specify the interface to use to gleam information, default: '$INTERFACE'"
    echo "-p, --port <number>       port number on which to contact the configuration server, default: '$PORT'"
    echo "-U, --uri <path>          URI path on which to contact the configuraiton server, default: '$URI'"
}

while [ $# -gt 0 ]; do
  case $1 in
    -h|--help)
      usage
      exit 0
      ;;
    -i|--interface)
      shift
      if [ $# -eq 0 ]; then
        echo "[error] must specify an interface name to use"  >&2
        usage
        exit 2
      fi
      INTERFACE=$1
      ;;
    -p|--port)
      shift
      if [ $# -eq 0 ]; then
        echo "[error] must specify a port number to use" >&2
        usage
        exit 2
      fi
      PORT=$1
      ;;
    -u|--uri)
      shift
      if [ $# -eq 0 ]; then
        echo "[error] must specify a URI to use" >&2
        exit 2
      fi
      URI=$1
      ;;
    *)
      echo "[error] unknown command line option '$1'" >&2
      usage
      exit 0
      ;;
  esac
  shift
done

INITIALIZATION="/var/lib/maas/dhcp/initialize"
INITIALIZATION_LOG="/var/lib/maas/dhcp/initialize.log"

INITIALIZATION="/tmp/init"
INITIALIZATION_LOG="/tmp/init.log"
MAC=`ifconfig $INTERFACE 2>/dev/null | grep HWaddr | awk '{print $5}'`

if [ "$MAC x" == " x" ]; then
  echo "[error] unable to find MAC address for specified interface '$INTERFACE'" >&2
  exit 2
fi

# Grab the IP address (service identifier) from every interface from which a DHCP lease was obtained
LEASE="/var/lib/dhcp/dhclient.$INTERFACE.leases"

if [ ! -f $LEASE ]; then
  echo "[error] unable to read DHCP lease file '$LEASE' for call home information" >&2
  exit 2
fi

SERVER=`grep -h dhcp-server-identifier $LEASE | awk '{print $3}' | tail -1 | sed -e 's/;//g'`

if [ "$SERVER x" == " X" ]; then
  echo "[error] unable to locate server address for interface '$INTERFACE'" >&2
  exit 2
fi

echo "[info] will call home to server '$SERVER' for registration and further configuration information"

BOOTTIME=`who -b | awk '{printf("%sT%s", $3, $4)}'`

# Continue to call home until we get a response (200 OK) from someone. There will be a backoff to 
# a limit so we don't continue to spam the network
INTERVAL=1
INCREMENT_FACTOR=2
MAX_INTERVAL=300
while true; do
  REQUEST="wget -t 5 -O $INITIALIZATION -S http://$SERVER:$PORT/$URI?mac=$MAC&boottime=$BOOTTIME"
  echo "[info] call home request: $REQUEST"
  RESULT=`$REQUEST 2>&1 | grep HTTP/ | grep "200 OK"`
  ERROR=$?
  if [ $ERROR -ne 0 ]; then
    echo "[error] call home request failed with error code '$ERROR', will attempt after pause of '$INTERVAL' seconds"
  else
    echo "[info] server response is: "`echo $RESULT | awk '{print $2 " " $3}'`
    if [ "$RESULT x" != " x" ]; then
      # Check if an initialization function was returned from the server
      SIZE=`stat -c %s $INITIALIZATION` 
      if [ $SIZE -eq 0 ]; then
        echo "[info] no intialization function returned from server"
        exit 0
      fi
      # have an initialization function, so make it executable and execute it
      chmod 755 $INITIALIZATION
      $INITIALIZATION 2>&1 >> $INITIALIZATION_LOG
      ERROR=$?
      if [ $ERROR -ne 0 ]; then
        echo "[error] initialization function returned an error code '$ERROR'" >&2
      fi
      exit $ERROR
    fi
  fi

  # Failed to connect to any server. Wait and try again
  sleep $INTERVAL

  # Increment interval for back off up to maxium
  INTERVAL=`expr $INTERVAL \* $INCREMENT_FACTOR`
  if [ $INTERVAL -gt $MAX_INTERVAL ]; then
    INTERVAL=$MAX_INTERVAL
  fi
done
