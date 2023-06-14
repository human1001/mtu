package rawnet

import (
	"net"
	"time"

	"github.com/google/gopacket/routing"
	"github.com/mdlayher/arp"
)

// GetSrcMAC 获取卡M默认网AC地址和吓一跳MAC地址
func GetMAC() (srcMAC, dstMAC net.HardwareAddr, err error) {
	r, err := routing.New()
	if err != nil {
		panic(err)
	}

	var ifi *net.Interface
	var getway net.IP
	ifi, getway, _, err = r.Route(net.IPv4zero)
	if err != nil {
		return
	}

	srcMAC = ifi.HardwareAddr

	var arpc *arp.Client
	if arpc, err = arp.Dial(ifi); err != nil {
		return
	} else {
		arpc.SetDeadline(time.Now().Add(time.Millisecond * 100))

		if dstMAC, err = arpc.Resolve(getway); err != nil {
			return
		} else {
			return
		}
	}
}
