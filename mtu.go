package mtu

var err error

type MTU struct {

	// PingHost PING命令请求地址, 默认baidu.com
	PingHost string
	// SeverAddr  IP或域名, 探测上行MTU设置
	SeverAddr string
	// Port 使用端口(UDP), 探测下行链路设置, 默认 19986
	Port int
}

func NewMTU(f func(m *MTU) *MTU) *MTU {

	var m = new(MTU)
	m = f(m)
	if m.PingHost == "" {
		m.PingHost = "baidu.com"
	}
	if m.Port == 0 {
		m.Port = 19986
	}
	return m
}

// Client 客户端
// 当 fastMode = true时, 将采用命令行告知的MTU, 如：ping: local error: message too long, mtu=1400
func (m *MTU) Client(isUpLink bool, fastMode bool) (uint16, error) {

	if isUpLink {
		//Uplink ping
		return clientUpLink(m.PingHost, fastMode)
	} else {
		//Downlink
		return clientDownLink(m.SeverAddr, m.Port)
	}
}

// Sever 服务, 探测下行链路需要
//  需要root权限
func (m *MTU) Sever() error {

	return m.sever()
}
