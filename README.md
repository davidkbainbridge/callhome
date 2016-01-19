# Call Home
Provides the ability for a device, such as a ONL switch or computer resource, to
register with and receive additional configuration from a configuration server.

# Process
This project provides a script and is run as an init.d script that is activated
when the system boots. This script attempts to locate a configuration server
from the DHCP lease information (i.e., the DHCP server today, may move to the
saddr server over time) and sends and HTTP GET request to that server sending
the clients MAC address as well as the boot time of the server. 

If the configuration server returns data to that request the script attempts
to execute that data as an executable (or script) and thus perform additional
post boot configuration actions.

The script will contine to send the request at intervals, with back off, until
it receives a response from the server. 
