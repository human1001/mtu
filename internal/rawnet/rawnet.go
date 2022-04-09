package rawnet

import (
	"net"
	"strings"
	"time"

	"github.com/google/gopacket/routing"
	"github.com/mdlayher/arp"
)

var err error

//  checkSum check sum
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

// PackageUDP package udp
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

// GetLocalIP 获取内网IP
func GetLocalIP() net.IP {
	var conn net.Conn
	if conn, err = net.Dial("udp", "114.114.114.114:80"); err != nil {
		return nil
	}
	defer conn.Close()
	return net.ParseIP(strings.Split(conn.LocalAddr().String(), ":")[0])
}

// SendIPPacketDFUDP send DF IP(UDP) packet
func SendIPPacketDFUDP(lIP, rIP net.IP, lPort, rPort uint16, d []byte) error {

	uR := PackageUDP(lIP, rIP, lPort, rPort, d)

	return sendIPPacketDFUDP(lIP, rIP, lPort, rPort, uR)
}

// GetSrcMAC 获取默认网卡MAC地址和吓一跳MAC地址
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
