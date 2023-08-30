package mtu

import (
	"errors"

	"github.com/human1001/mtu/internal/com"
)

var err error

type MTU struct {
	PingHost  string // 探测上行MTU时设置, PING命令请求地址, 默认baidu.com
	SeverAddr string // 下行MTU时设置, 服务器IP或域名

	// Port 使用端口(UDP), 探测下行链路时设置, 默认 19986
	Port int
}

// NewMTU
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
//
//	isUpLink 探测上行路径MTU，否则探测下行MTU
//	fastMode 探测上行链路的快速模式，更快
//	 如数据包太大时,Ubuntu会提示：ping: local error: message too long, mtu=1400，快速模式直接采用1400这个值
func (m *MTU) Client(isUpLink bool, fastMode bool) (uint16, error) {

	if isUpLink {
		//Uplink ping
		return clientUpLink(m.PingHost, fastMode)
	} else {
		//Downlink
		if m.SeverAddr == "" {
			return 0, errors.New("must set MTU.SeverAddr")
		}
		if ip, err := com.ToIP(m.SeverAddr); err != nil {
			return 0, err
		} else {
			m.SeverAddr = ip.String() //最终SeverAddr的是IP
		}
		return clientDownLink(m.SeverAddr, m.Port)
	}
}

// Sever 服务, 探测下行链路需要
//
//	需要发送自定义IP包, 需要root权限运行。
//	确保服务器的上行MTU足够大, 不应成为链路瓶颈
func (m *MTU) Sever() error {

	return m.sever()
}
