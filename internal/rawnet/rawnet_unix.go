// +build linux netbsd openbsd darwin freebsd

package rawnet

import (
	"errors"
	"net"

	"golang.org/x/net/ipv4"
)

// Protocol https://zh.wikipedia.org/wiki/IP协议号列表
// 1  ICMP
// 4  IPv4
// 6  TCP
// 17 UDP
// 41 IPv6

// PackageIPHeader package ip header
func PackageIPHeader(lIP, rIP net.IP, lport, rport uint16, flag, offset int, protocol uint8, d []byte) (*ipv4.Header, error) {

	// refer https://zh.wikipedia.org/wiki/IPv4
	// refer https://zhangbinalan.gitbooks.io/protocol/content/ipxie_yi_tou_bu.html

	iph := &ipv4.Header{
		Version:  ipv4.Version,
		Len:      ipv4.HeaderLen,
		TOS:      0x00,
		TotalLen: ipv4.HeaderLen + len(d),
		TTL:      64,
		Flags:    flag,
		FragOff:  offset,
		Protocol: int(protocol),
		Checksum: 0,
		Src:      lIP,
		Dst:      rIP,
	}

	h, err := iph.Marshal()
	if err != nil {
		return nil, err
	}
	iph.Checksum = int(checkSum(h))
	return iph, nil
}

// sendIPPacket except windows system
func sendIPPacket(lIP, rIP net.IP, lPort, rPort uint16, flag, offset int, protocol uint8, d []byte) error {

	udpPack := PackageUDP(lIP, rIP, lPort, rPort, d)
	iph, err := PackageIPHeader(lIP, rIP, lPort, rPort, flag, offset, protocol, udpPack)
	if err != nil {
		return err
	}

	ipAddr, err := net.ResolveIPAddr("ip4", lIP.String())
	if err != nil {
		return err
	}
	conn, err := net.ListenIP("ip4:udp", ipAddr)
	if err != nil {
		return err
	}
	rawConn, err := ipv4.NewRawConn(conn) //ipconn to rawconn
	if err != nil {
		return err
	}

	err = rawConn.WriteTo(iph, udpPack, nil)
	if err != nil {
		return err
	}
	return nil
}

// sendIPPacketDFUDP
func sendIPPacketDFUDP(lIP, rIP net.IP, lPort, rPort uint16, d []byte) error {

	raddr, err1 := net.ResolveIPAddr("ip4:udp", rIP.String())
	laddr, err2 := net.ResolveIPAddr("ip4:udp", lIP.String())
	con, err3 := net.DialIP("ip4:udp", laddr, raddr)
	if err1 != nil || err2 != nil || err3 != nil {
		return errors.New(err1.Error() + err2.Error() + err3.Error())
	}
	defer con.Close()

	err := sendIPPacket(lIP, rIP, lPort, rPort, 2, 0, 17, d)
	return err
}
