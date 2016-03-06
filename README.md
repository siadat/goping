A simple implementation of ping by sending and receiving ICMP echo packets.

Usage:

	goping <ip> [-c count] [-i interval]

Example:

    sudo ./goping 8.8.8.8
    2016/03/07 00:32:48 GOPING 18 bytes to 8.8.8.8
    2016/03/07 00:32:49 #1 18 bytes from 8.8.8.8 479.674743ms
    2016/03/07 00:32:51 #2 18 bytes from 8.8.8.8 505.850823ms
    2016/03/07 00:32:53 #3 18 bytes from 8.8.8.8 633.294553ms
