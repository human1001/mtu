package rawnet

import (
	"errors"
	"net"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/net/ipv4"
)

// 原生net socke，在传输层和网络层

/*
* 传输层 传输允许自定义设置传输包头，如UDP、ICMP，不允许设置网络层包、即IP包
 */

//  checkSum 校验和
func checkSum(d []byte) uint16 {
	var S uint32
	l := len(d)
	if l&0b1 == 1 { //奇数
		for i := 0; i < l-1; {
			S = S + uint32(d[i])<<8 + uint32(d[i+1])
			if S>>16 > 0 { // 反码加法 溢出加一
				S = S&0xffff + 1
			}

			i = i + 2
		}
		S = S + uint32(d[l-1])<<8
	} else {
		for i := 0; i < l; {
			S = S + uint32(d[i])<<8 + uint32(d[i+1])
			if S>>16 > 0 { // 反码加法 溢出加一
				S = S&0xffff + 1
			}

			i = i + 2
		}
	}

	return uint16(65535) - uint16(S)
}

// PackageUDP 打包为UDP包
func PackageUDP(laddr, raddr net.IP, lport, rport uint16, d []byte) []byte {
	// 参考 https://zh.wikipedia.org/wiki/用户数据报协议
	// UDP包不需要IP，形参需要地址是用于构成伪包、计算校验和
	var P []byte = make([]byte, 20, len(d)+20)
	//伪头
	P[0], P[1], P[2], P[3] = laddr[12], laddr[13], laddr[14], laddr[15]
	P[4], P[5], P[6], P[7] = raddr[12], raddr[13], raddr[14], raddr[15]
	P[8], P[9] = 0, 17 //协议类型UDP
	// 实头
	P[10], P[11] = uint8((len(d)+8)>>8), uint8(len(d)+8) //整个包长度
	P[12], P[13] = uint8(lport>>8), uint8(lport)         //源端口
	P[14], P[15] = uint8(rport>>8), uint8(rport)
	P[16], P[17] = uint8((len(d)+8)>>8), uint8(len(d)+8) //长度
	P[18], P[19] = 0, 0                                  //校验和

	P = append(P, d...)

	rS := checkSum(P)
	P[18], P[19] = uint8(rS>>8), uint8(rS) //校验和(包括伪头、不包括数据)

	return P[12:] //不包括伪头
}

// Protocol https://zh.wikipedia.org/wiki/IP协议号列表
// 1  ICMP
// 4  IPv4
// 6  TCP
// 17 UDP
// 41 IPv6

// GetLocalIP Get LAN IPv4
func GetLocalIP() net.IP {
	raddr, err1 := net.ResolveUDPAddr("udp4", "120.120.120.120:438")
	laddr, err2 := net.ResolveUDPAddr("udp4", ":")
	con, err3 := net.DialUDP("udp4", laddr, raddr)
	if err1 != nil || err2 != nil || err3 != nil {
		return nil
	}
	defer con.Close()
	return net.ParseIP(strings.Split(con.LocalAddr().String(), ":")[0])
}

// PackageIPHeader package ip header
func PackageIPHeader(lIP, rIP net.IP, lport, rport uint16, flag, offset int, protocol uint8, d []byte) (*ipv4.Header, error) {

	// 参考 https://zh.wikipedia.org/wiki/IPv4
	// 参考 https://zhangbinalan.gitbooks.io/protocol/content/ipxie_yi_tou_bu.html

	var f ipv4.HeaderFlags
	if offset == 2 { //don't Fragment
		f = ipv4.DontFragment
		if offset != 0 {
			err := errors.New("header flag and fragment conflict")
			return nil, err
		}
	} else if offset == 1 {
		f = 1
	} else if offset == 0 { // END
		f = 0
	} else {
		err := errors.New("incorrect flag")
		return nil, err
	}

	iph := &ipv4.Header{
		Version:  ipv4.Version,
		Len:      ipv4.HeaderLen,
		TOS:      0x00,
		TotalLen: ipv4.HeaderLen + len(d),
		TTL:      64,
		Flags:    f,
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
	//计算IP头部校验值
	iph.Checksum = int(checkSum(h))
	return iph, nil
}

// SendIPPacket except windows system
func SendIPPacket(lIP, rIP net.IP, lPort, rPort uint16, flag, offset int, protocol uint8, d []byte) error {

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

// SendIPPacketDF send DF IP Packet(UDP),need root authority
func SendIPPacketDF(lIP, rIP net.IP, lPort, rPort uint16, d []byte) error {

	uR := PackageUDP(lIP, rIP, lPort, rPort, d)

	// sendmsg: not implemented on windows/amd64
	if runtime.GOOS == "windows" {
		// return rawnetwindows.SendIPPacketDF(rIP, rPort, 17, uR)
		_, p, n, _ := runtime.Caller(0)
		err := errors.New("Please annotate" + p + " " + strconv.Itoa(n) + "-" + strconv.Itoa(n+3) + "lines; and uncomment " + strconv.Itoa(n-1) + " line")
		return err
	} else if runtime.GOOS == "linux" {

		raddr, err1 := net.ResolveIPAddr("ip4:udp", rIP.String())
		laddr, err2 := net.ResolveIPAddr("ip4:udp", lIP.String())
		con, err3 := net.DialIP("ip4:udp", laddr, raddr)
		if err1 != nil || err2 != nil || err3 != nil {
			return errors.New(err1.Error() + err2.Error() + err3.Error())
		}
		defer con.Close()

		err := SendIPPacket(lIP, rIP, lPort, rPort, 2, 0, 17, uR)
		return err
	} else {
		err := errors.New("the OS " + runtime.GOOS + " not bu support")
		return err
	}
}
