package rawnet

import (
	"net"
	"net/netip"
	"time"

	"github.com/google/gopacket/routing"
	"github.com/mdlayher/arp"
)

// GetMAC 取得网卡与下一条网卡
func GetMAC() (srcMAC, dstMAC net.HardwareAddr, err error) {

	r, err := routing.New()
	if err != nil {
		panic(err)
	}

	var ifi *net.Interface
	var gateway net.IP
	ifi, gateway, _, err = r.Route(net.IPv4zero)
	if err != nil {
		return
	}

	srcMAC = ifi.HardwareAddr

	var arpc *arp.Client
	if arpc, err = arp.Dial(ifi); err != nil {
		return
	} else {
		err = arpc.SetDeadline(time.Now().Add(time.Millisecond * 100))
		if err != nil {
			return
		}
		addrFromIp, _ := netip.AddrFromSlice(gateway)
		if dstMAC, err = arpc.Resolve(addrFromIp); err != nil {
			return
		} else {
			return
		}
	}
}
