###### 上行链路MTU发现策略

使用UDP数据包

| 发送者       | 接受者       | 数据             | 说明                                                         |
| ------------ | ------------ | ---------------- | ------------------------------------------------------------ |
| client:19989 | sever:19989  | Muuid:a          | 探测下行MTU，sever接到此数据包后将发送DF IP(UDP)包(b)和c     |
| sever:19989  | client:19989 | Muuid: b:....    | 发送的DF IP(UDP)包                                           |
| sever:19989  | client:19989 | Muuid:c:len:step | 在b之后延时30ms发送，表示sever发送了长度为len+28(udp header+ip header)的DF IP数据包，不能收到则判定大于mtu(d)，能收到则发送更大的数据包(e) |
| client:19989 | sever:19989  | Muuid:d:len:step | sever收到此数据包将回复len-step长度的UDP(DF)包               |
| client:19989 | sever:19989  | Muuid:e:len:step | sever收到此数据包将回复len+step长度的UDP(DF)包               |

说明：

​		数据栏中`...`表示任意数据；`:`实际不存在，便于阅读用。client回复的step(d、e)等于收到的step(c)整除以2。当client收到的step等于1时可得出mtu。len和step各占两个字节。

​		Muuid为唯一id标识，无必要、属设计缺陷，但被保留。

