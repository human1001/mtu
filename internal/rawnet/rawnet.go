package rawnet

import (
	"errors"
	"fmt"
	"net"
	"seemfast/internal/com"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
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

// PackageICMP 打包为ICMP包
func PackageICMP(t, c uint8, d []byte) []byte {
	// 参数含义 参考 https://zh.wikipedia.org/zh-hans/互联网控制消息协议

	var P []byte = make([]byte, 4, 8+len(d)+1)
	P[0] = t                  //type
	P[1] = c                  //code
	P = append(P, 0, 0, 0, 0) //Rest of Header

	P = append(P, d...)
	var add bool = false
	if len(P)&1 != 0 {
		add = true
		P = append(P, uint8(0))
	}
	var rS uint16 = 0
	rS = checkSum(P) // ICMP的校验和包括数据
	P[2] = uint8(rS >> 8)
	P[3] = uint8(rS % 256)
	if add {
		P = P[:len(P)-1]
	}
	return P
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

// t 传输层 发送实例
func t() {
	raddr, err1 := net.ResolveIPAddr("ip4:icmp", "220.181.38.148")
	laddr, err2 := net.ResolveIPAddr("ip4:icmp", "")
	con, err3 := net.DialIP("ip4:icmp", laddr, raddr)
	defer con.Close()
	fmt.Println(con.LocalAddr().String())

	sR := PackageICMP(8, 0, []byte("safweofwefoiw"))
	// sR := PackageUDP(net.ParseIP("192.168.43.183"), net.ParseIP("220.181.38.148"), 19986, 19986, []byte("safweofwefoiw"))
	n, err4 := con.Write(sR)
	com.Errorlog(err1, err2, err3, err4)

	fmt.Println(n)
}

/*
* 网络层，在网络层可以设置IP包包头
* ipv4包可实现对非Windows的IP包的Raw读写，需要root权限
* 对于Windows可以调用/sys/windows包，但只能设置部分参数，无需管理员权限
 */
// 常见Protocol https://zh.wikipedia.org/wiki/IP协议号列表
// 1  ICMP
// 4  IPv4
// 6  TCP
// 17 UDP
// 41 IPv6

// GetLocalIP 获取局域网IPv4
func GetLocalIP() net.IP {
	raddr, err1 := net.ResolveUDPAddr("udp4", "120.120.120.120:438")
	laddr, err2 := net.ResolveUDPAddr("udp4", ":")
	con, err3 := net.DialUDP("udp4", laddr, raddr)
	if com.Errorlog(err1, err2, err3) {
		return nil
	}
	defer con.Close()
	return net.ParseIP(strings.Split(con.LocalAddr().String(), ":")[0])
}

// PackageIPHeader 计算IP包头
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
	if com.Errorlog(err) {
		return nil, err
	}
	//计算IP头部校验值
	iph.Checksum = int(checkSum(h))
	return iph, nil
}

//PackageIPPacket 打包为IP包
// 与PackageIPHeader不同，返回完整IP包
func PackageIPPacket(lIP, rIP net.IP, lport, rport uint16, protocol uint8, d []byte) []byte {
	if len(d) > 1372 { //MTU 1372
		err := errors.New("IP package's data too long, over 1372")
		com.Errorlog(err)
	}

	var h []byte = make([]byte, 0, 20)
	h = append(
		h,
		69,                    // 69 = 0b0100 0101 => IPv4,IP header长度为4*5 = 20B
		0,                     // 区分服务 和 显式拥塞通告
		uint8((len(d)+20)>>8), // 整个IP包全长
		uint8(len(d)+20),
		0,    //标识符
		0,    //标识符
		2<<5, //标志和分片偏移 010 DF
		0,    //标志和分片偏移 010 DF
		128,  // TTL
		protocol,
		0,       // checkSum (切片下标10)
		0,       // checkSum (切片下标11)
		lIP[12], //源IP
		lIP[13],
		lIP[14],
		lIP[15],
		rIP[12], //目的IP
		rIP[13],
		rIP[14],
		rIP[15],
	)
	cS := checkSum(h)

	h[10] = uint8(cS >> 8)
	h[11] = uint8(cS)

	d = append(h, d...)
	return d
}

// SendIPPacket 发送IP包，允许对IP包进行常规设置(非Windows且ROOT权限)
func SendIPPacket(lIP, rIP net.IP, lPort, rPort uint16, flag, offset int, protocol uint8, d []byte) error {

	udpPack := PackageUDP(lIP, rIP, lPort, rPort, d)
	iph, err := PackageIPHeader(lIP, rIP, lPort, rPort, flag, offset, protocol, udpPack)
	if com.Errorlog(err) {
		return err
	}

	ipAddr, err := net.ResolveIPAddr("ip4", lIP.String())
	if com.Errorlog(err) {
		return err
	}
	conn, err := net.ListenIP("ip4:udp", ipAddr)
	if com.Errorlog(err) {
		return err
	}
	rawConn, err := ipv4.NewRawConn(conn) //ipconn to rawconn
	if com.Errorlog(err) {
		return err
	}

	err = rawConn.WriteTo(iph, udpPack, nil)
	if com.Errorlog(err) {
		return err
	}
	return nil
}

//t2 接收IP包，非Windows，ROOT权限
func t2() error {
	// 以后写
	c, err := net.ListenPacket("ip4:icmp", "0.0.0.0") // OSPF for IPv4
	if com.Errorlog(err) {
		return err
	}
	defer c.Close()
	ipconn, err := ipv4.NewRawConn(c)
	if com.Errorlog(err) {
		return err
	}
	for {
		buf := make([]byte, 1480)
		hdr, payload, controlMessage, err := ipconn.ReadFrom(buf)
		if err != nil {
			fmt.Println(5, err)
			return nil
		}
		if payload != nil {
			fmt.Println(hdr, controlMessage)
			fmt.Println(payload)
		}
	}
}

/*
* 数据链路层
 */
// getDevice 得到局域网IP所在的网卡名

// GetDevice 获取网卡名
func GetDevice(ip string) string {
	devices, err := pcap.FindAllDevs()
	if com.Errorlog(err) {
		return ""
	}
	for _, v := range devices {
		for _, w := range v.Addresses {
			if w.IP.String() == ip {
				return v.Name
			}
		}
	}

	return ""
}

// GetSrcMAC 获得局域网IP所在网卡的MAC
func GetSrcMAC(ip string) net.HardwareAddr {
	rl, err := net.Interfaces()
	if com.Errorlog(err) {
		fmt.Println(err)
		return nil
	}
	for _, v := range rl {
		addrs, err := (&v).Addrs()
		if com.Errorlog(err) {
			fmt.Println(err)
			return nil
		}
		for _, addr := range addrs {
			if strings.Contains(addr.String(), ip) {
				return v.HardwareAddr
			}
		}

	}

	return nil
}

// GetDstMAC 获取目的MAC(Linux root authority)
func GetDstMAC(localIP, device string) net.HardwareAddr {
	// 通过cpac捕获以太包获取dstMAC，当协程发包失败且本地端口没有通信时会长时间阻塞
	// 正常情况此函数耗时1ms

	var (
		snapshotLen int32 = 65536
		promiscuous bool  = true
		err         error
		timeout     time.Duration = time.Microsecond
		handle      *pcap.Handle

		linkLayer    gopacket.LinkLayer
		networkLayer gopacket.NetworkLayer
	)

	// Open device
	handle, err = pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
	if com.Errorlog(err) {
		return nil
	}
	defer handle.Close()
	go func() {
		time.Sleep(time.Microsecond * 5)
		laddr, err1 := net.ResolveUDPAddr("udp4", "")
		raddr, err2 := net.ResolveUDPAddr("udp4", "120.120.120.120:439")
		uh, err3 := net.DialUDP("udp4", laddr, raddr)
		if com.Errorlog(err1, err2, err3) {
			return
		}
		defer uh.Close()
		uh.Write([]byte("lysShublysShublysShub"))
	}()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		linkLayer = packet.LinkLayer()
		networkLayer = packet.NetworkLayer()
		if networkLayer == nil || linkLayer == nil {
			continue
		}
		ipHeader := networkLayer.LayerContents()
		linkHeader := linkLayer.LayerContents()
		var lhs []string
		for _, v := range linkHeader {
			lhs = append(lhs, fmt.Sprintf("%02X", v))
		}
		if len(ipHeader) == 20 {
			if localIP == (net.IP{ipHeader[12], ipHeader[13], ipHeader[14], ipHeader[15]}).String() { //send
				if len(linkHeader) == 14 { //前7
					macStr := lhs[0] + ":" + lhs[1] + ":" + lhs[2] + ":" + lhs[3] + ":" + lhs[4] + ":" + lhs[5]
					hw, err := net.ParseMAC(macStr)
					if com.Errorlog(err) {
						continue
					}
					return hw
				}
			} else if localIP == (net.IP{ipHeader[16], ipHeader[17], ipHeader[18], ipHeader[19]}).String() { //receive
				if len(linkHeader) == 14 { //后7
					macStr := lhs[6] + ":" + lhs[7] + ":" + lhs[8] + ":" + lhs[9] + ":" + lhs[10] + ":" + lhs[11]
					hw, err := net.ParseMAC(macStr)
					if com.Errorlog(err) {
						continue
					}
					return hw
				}
			}
		}
	}
	return nil
}

// SendEtherPacket 发送以太包，可对IP进行非常规操作、如伪造源IP
// d为完整IP包数据
func SendEtherPacket(device string, dstMAC, srcMAC net.HardwareAddr, d []byte) bool {
	// Open device
	handle, err := pcap.OpenLive(device, 1500, true, time.Microsecond)
	if com.Errorlog(err) {
		return false
	}
	defer handle.Close()

	dstMAC = append(dstMAC, srcMAC...)
	dstMAC = append(dstMAC, 8, 0) //type:0x0800
	d = append(dstMAC, d...)

	err = handle.WritePacketData(d)
	if com.Errorlog(err) {
		return false
	}
	return true
}

/*
* 功能函数
 */

// SendForgeSrcIPUDP 发送伪造源IP的UDP数据包
// rIP:目的IP fIP:伪造的源IP
// 需root权限，NAT会更改源IP，严格路由会被丢弃(fIP尽量同网段)
func SendForgeSrcIPUDP(rIP, fIP net.IP, lport, rport uint16, d []byte) bool {
	// laddr, err1 := net.ResolveUDPAddr("udp", "")
	// raddr, err2 := net.ResolveUDPAddr("udp", "47.102.124.216:19986")
	// conn, err3 := net.DialUDP("udp", laddr, raddr)
	// if com.Errorlog(err1, err2, err3) {
	// 	return
	// }
	// n, err1 := conn.Write([]byte("sdfasfdsfa"))
	// if com.Errorlog(err1) {
	// 	return
	// }
	// fmt.Println(n)
	// fmt.Println(rawnet.SendForgeSrcIPUDP(net.ParseIP("47.102.124.216"), net.ParseIP("192.168.0.220"), 19986, 19986, []byte("sdfasfdsfa")))

	if len(d) > 1364 {
		return false
	}

	lIP := GetLocalIP()
	if lIP == nil {
		return false
	}

	// 打包为UDP
	d = PackageUDP(fIP, rIP, lport, rport, d)

	// ih, err1 := PackageIPHeader(fIP, rIP, lport, rport, 2, 0, 17, d) //DF UDP
	// ihb, err2 := ih.Marshal()
	// if com.Errorlog(err1, err2) {
	// 	return false
	// }
	// d = append(ihb, d...)
	d = PackageIPPacket(fIP, rIP, lport, rport, 17, d)

	device := GetDevice(lIP.String())
	if device == "" {
		return false
	}
	dstMAC := GetDstMAC(lIP.String(), device)
	srcMAC := GetSrcMAC(lIP.String())

	if SendEtherPacket(device, dstMAC, srcMAC, d) {
		return true
	}
	return false
}

// SendDFIPPacket only linux and root authority
func SendDFIPPacket(lIP, rIP net.IP, lPort, rPort uint16, d []byte) bool {

	uR := PackageUDP(lIP, rIP, lPort, rPort, d)

	raddr, err1 := net.ResolveIPAddr("ip4:udp", rIP.String())
	laddr, err2 := net.ResolveIPAddr("ip4:udp", lIP.String())
	con, err3 := net.DialIP("ip4:udp", laddr, raddr)
	if com.Errorlog(err1, err2, err3) {
		return false
	}
	defer con.Close()

	err := SendIPPacket(lIP, rIP, lPort, rPort, 2, 0, 17, uR)
	if com.Errorlog(err) {
		return false
	}

	return true
}
