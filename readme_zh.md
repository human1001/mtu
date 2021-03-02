# MTUdiscover

发现链路的MTU

##### 注意：

此项目返回的MTU是UDP数据包数据的最大大小；如返回1372，则IP包最大不分包大小为：1372 + 8(UDP包头) + 20 (IP包头) = 1400

如果在Go项目中使用，请克隆到本地作为私有包使用，请勿直接在代码中`import`,因为你需要修改源码；如果不在Go项目中使用，可以打包为动态链库(C-share)嵌入到几乎所有的项目中。

探测上行链路是使用的`PING`命令发送DF包，只针对几种常见的系统实现，其他的你可以自己补充。探测下行链路你需要服务器支持(root权限)

**开始**

- `git clone https://github.com/lysShub/MTUdiscover.git`
- `cd ./MTUdiscover/test`
- `go run main.go`

可以得到上行链路的MTU。

