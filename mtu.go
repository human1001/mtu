package mtu

import (
	"errors"
	"mtu/internal/com"
	"mtu/internal/rawnet"
	"net"
	"strconv"
)

// Discover the MTU of the link by UDP packet

// sever sever addr, ip or domain
const sever string = "114.116.254.26"

// port port used by the server and client
const port uint16 = 19989

// pingHost ping host
const pingHost string = "baidu.com"

// Client client
// if isUpLink = false, it will discover downlink's mtu, need sever support
// discover the uplink through the PING command
// may block for ten seconds; for example, PING command didn't replay
func Client(isUpLink bool, UpLinkFast bool) uint16 {

	if isUpLink {
		//Uplink ping

		return com.ClientUpLink(pingHost, UpLinkFast)
	}

	//Downlink
	return com.ClientDownLink(sever, port)

}

// Sever sever need root authority, remember open UDP port
// detect downlink MTU by sending IP(DF) packets
//
func Sever() error {

	laddr, err1 := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(int(port)))
	handle, err2 := net.ListenUDP("udp", laddr)
	if err1 != nil || err2 != nil {
		// log error
		return errors.New(err1.Error() + err2.Error())
	}
	defer handle.Close()

	lIP := rawnet.GetLocalIP()
	if lIP == nil {
		err := errors.New("can't get local IP or network card name")
		// log error
		return err
	}

	var get bool = false
	var bodyB, bodyC []byte
	for {
		d := make([]byte, 2000)
		get = false

		n, raddr, err := handle.ReadFromUDP(d)
		if err != nil {
			// log error
		}
		bodyB, bodyC = make([]byte, 37), make([]byte, 37)
		copy(bodyB, d)
		copy(bodyC, d)
		bodyB = append(bodyB, 0xb)

		if n == 38 && d[37] == 0xa { //get a

			bodyB = append(bodyB, make([]byte, 962)...) //962 + 38 = 1000
			bodyC = append(bodyC, 0xc, 3, 232, 3, 232)  //len,step=1000
			get = true
		} else if n == 42 && d[37] == 0xd { // get d

			len := int(d[38])<<8 + int(d[39])
			step := int(d[40])<<8 + int(d[41])
			len = len - step
			if len < 38 {
				len = 38
			}
			bodyB = append(bodyB, make([]byte, len-38)...)
			bodyC = append(bodyC, 0xc, uint8(len>>8), uint8(len), d[40], d[41])
			get = true
		} else if n == 42 && d[37] == 0xe { //get e

			len := int(d[38])<<8 + int(d[39])
			step := int(d[40])<<8 + int(d[41])
			len = len + step

			bodyB = append(bodyB, make([]byte, len-38)...)
			bodyC = append(bodyC, 0xc, uint8(len>>8), uint8(len), d[40], d[41])
			get = true
		}

		if get {
			_, err := handle.WriteToUDP(bodyC, raddr) //reply c
			if err != nil {
				// log error
			}
			err = rawnet.SendIPPacketDFUDP(lIP, raddr.IP, port, uint16(raddr.Port), bodyB) //reply b
			if err != nil {
				// log error
			}
		}

	}
}
