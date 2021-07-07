# MTUdiscover

用于发现链路的MTU

- 此项目返回的值是UDP数据包数据的大小；如返回1372，则IP包最大不分包大小为：1372 + 8(UDP包头) + 20 (IP包头) = 1400
- 探测上行链路是使用的`PING DF`命令发送数据包，支持常见系统。探测下行链路采用二分法发现MTU
- 探测下行链路需要服务器支持

###### 快速开始

​		[参考](https://github.com/lysShub/mtu/blob/master/test/main.go)

