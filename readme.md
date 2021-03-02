# MTUdiscover

[【中文】](https://github.com/lysShub/MTUdiscover/blob/master/readme_zh.md)

##### NOTE:

**<font color="red">The MTU obtained in this project refers to the maximum allowable data of udp packets</font>**

​		If we get MTU = 1350 by this project, maximum IP packet size:

`	1350 + 8(udp header) + 20(IPv4 header) = 1378 Bytes `



##### Description:

​		Uplink and downlink have different MTU；discover uplink uses PING command,discover downlink need a sever. Found that the uplink mtu has requirements for the system, <font color="yellow">I only test on Android, Linux, Windows  </font>. protocol **PMTUD** is not widely supported, so it is not used.

​		You should `git Clone https://github.com/lysShub/MTUdiscover.git` instead  `import ("github.com/lysShub/MTUdiscover")` in your codes. Because you need to modify the source code:

- /mtu.go->port const
- if your sever OS is Windows, channel annotate `return rawnetwindows.SendIPPacketDF(rIP, rPort, 17, uR)` at /internal/rawnet/rawnet.go->SendIPPacketDF func
- If you are not using under Windows, Android,Linux system, add codes about at mtu.go 156 lines.

##### start:

​	**condition:** Golang and GO111MODULE=on

​	**get uplink mtu:**

- `git clone https://github.com/lysShub/MTUdiscover.git`
- `cd ./MTUdiscover/test`
- `go run main.go`

