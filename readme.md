# MTUdiscover

**<font color="red">NOTE：The MTU obtained in this project refers to the maximum allowable data of udp packets</font>**

If we get MTU = 1350 by this project, maximum IP packet size:

​	1350 + 8(udp header) + 20(IPv4 header) = 1378 Bytes

Uplink and downlink have different MTU, discover downlink need a sever.

Found that the uplink mtu has requirements for the system, <font color="yellow">I only test on Android, Linux, Windows and </font>

protocol **PMTUD** is not widely supported, so it is not used