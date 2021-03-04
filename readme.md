# MTUdiscover

[【中文】](https://github.com/lysShub/MTUdiscover/blob/master/readme_zh.md)

##### NOTE:

**<font color="red">The MTU obtained in this project refers to the maximum allowable data of udp packets</font>**

​		If we get MTU = 1350 by this project, maximum IP packet size:

`	1350 + 8(udp header) + 20(IPv4 header) = 1378 Bytes `



##### Description:

​		Uplink and downlink have different MTU；discover uplink uses PING command, discover downlink need a sever.  protocol **PMTUD** is not widely supported, so it is not used.

​		You should `git Clone https://github.com/lysShub/MTUdiscover.git` instead  `import ("github.com/lysShub/MTUdiscover")` in your codes. Because you need to modify the source code, e.g:

- port
- sever
- pingHost

##### start:

​	Golang >=go1.15 ;  GO111MODULE=on

​	**get uplink mtu:**

- `git clone https://github.com/lysShub/MTUdiscover.git`
- `cd ./MTUdiscover/test`
- `go run main.go`

