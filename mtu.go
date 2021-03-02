package mtu

import (
	"bytes"
	"errors"
	"mtu/internal/com"
	"mtu/internal/rawnet"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Discover the MTU of the link by UDP packet

// sever sever addr, ip or domain
const sever string = ""

// port port used by the server and client
const port uint16 = 19989

// pingHost ping host
const pingHost string = "baidu.com"

// Client client
// if isUpLink = false, it will discover downlink's mtu, need sever support
// discover the uplink through the PING command
// may block for ten seconds; for example, PING command didn't replay
func Client(isUpLink bool) (uint16, error) {

	if isUpLink { //Uplink ping
		var d []byte
		var li [][]byte
		var n int
		var cmd *exec.Cmd
		var stdout, stderr bytes.Buffer
		var wrap []byte = make([]byte, 0)
		var F func(l int) int // 0:error or exception,eg:timeout; -1:too small  1:too big

		if runtime.GOOS == "windows" {
			//  Used to split blank lines
			wrap = append(wrap, []byte{13, 10, 13, 10}...) // windows Wrap：\r\n=Enter(13) Wrap(10)
			F = func(l int) int {
				cmd = exec.Command("cmd", "/C", "ping", "-f", "-n", "1", "-l", strconv.Itoa(l), "-w", "1000", pingHost)
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr

				err := cmd.Run()
				d = make([]byte, 0)
				if err != nil {
					d = append(d, []byte(err.Error())...)
				}
				d = append(d, stderr.Bytes()...)
				d = append(d, stdout.Bytes()...)
				d = com.ToUtf8(d)

				li = bytes.Split(d, wrap)
				for _, v := range li {
					rv := string(v)
					if strings.Contains(rv, strconv.Itoa(l)) {
						if strings.Contains(rv, `TTL`) { //small
							return -1
						} else if strings.Contains(rv, `DF`) {
							return 1
						}
					}
				}
				return 0
			}
		} else if runtime.GOOS == "linux" {
			wrap = append(wrap, []byte{10, 10}...) // Linux Wrap：\n=Wrap
			F = func(l int) int {
				cmd = exec.Command("ping", "-M", "do", "-c", "1", "-s", strconv.Itoa(l), "-W", "1", pingHost)
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr
				err := cmd.Run()
				d = make([]byte, 0)
				if err != nil {
					d = append(d, []byte(err.Error())...)
				}
				d = append(d, stderr.Bytes()...)
				d = append(d, stdout.Bytes()...)
				d = com.ToUtf8(d[:n])

				li = bytes.Split(d, wrap)
				for _, v := range li {
					rv := string(v)
					if strings.Contains(rv, strconv.Itoa(l)) {
						if strings.Contains(rv, `ttl`) { //small
							return -1
						} else if strings.Contains(rv, `too long`) {
							return 1
						}
					}
				}
				return 0
			}
		} else if runtime.GOOS == "darwin" {
			wrap = append(wrap, []byte{13, 13}...) // macOS Wrap：\r=Enter
			F = func(l int) int {
				cmd = exec.Command("cmd", "/C", "ping", "-c", "1", "-t", "1", "-D", "-s", strconv.Itoa(l), pingHost)
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr
				err := cmd.Run()
				d = make([]byte, 0)
				if err != nil {
					d = append(d, []byte(err.Error())...)
				}
				d = append(d, stderr.Bytes()...)
				d = append(d, stdout.Bytes()...)
				d = com.ToUtf8(d[:n]) //编码

				li = bytes.Split(d, wrap)
				for _, v := range li {
					rv := string(v)
					if strings.Contains(rv, strconv.Itoa(l)) {
						if strings.Contains(rv, `ttl`) { //small
							return -1
						} else if strings.Contains(rv, `too long`) {
							return 1
						}
					}
				}
				return 0
			}
		} else if runtime.GOOS == "android" {
			wrap = append(wrap, []byte{10, 10}...) // Linux Wrap：\n=Wrap
			F = func(l int) int {
				cmd = exec.Command("/bin/ping", "-M", "do", "-c", "1", "-s", strconv.Itoa(l), "-W", "1", pingHost) //"-c", "1", pingHost
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr
				err := cmd.Run()
				d = make([]byte, 0)
				if err != nil {
					d = append(d, []byte(err.Error())...)
				}
				d = append(d, stderr.Bytes()...)
				d = append(d, stdout.Bytes()...)
				d = com.ToUtf8(d)

				li = bytes.Split(d, wrap)
				for _, v := range li {
					rv := string(v)
					if strings.Contains(rv, strconv.Itoa(l)) {
						if strings.Contains(rv, `ttl`) { //small
							return -1
						} else if strings.Contains(rv, `too long`) {
							return 1
						}
					}
				}
				return 0
			}
		} else {
			err := errors.New("system " + runtime.GOOS + " not be supported")
			return 0, err
		}

		// Binary search
		left, right, mid, step := 1, 2000, 0, 1999
		for {
			mid = int(float64((left + right) / 2))
			r := F(mid)

			if 1 == r { //big
				right = mid - 1
			} else if -1 == r { //small
				left = mid + 1
			} else { // r==0 error or exception
				break
			}
			step = right - left
			if step <= 3 {
				for i := right + 1; i <= left; i-- {
					if F(i) == -1 {
						return uint16(i), nil
					}
				}
			}
		}
	} else { //Downlink precision = 10

		raddr1, err1 := net.ResolveUDPAddr("udp", sever+":"+strconv.Itoa(int(port)))
		laddr1, err2 := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(int(port)))
		conn, err3 := net.DialUDP("udp", laddr1, raddr1)
		if err1 != nil || err2 != nil || err3 != nil {
			return 0, errors.New(err1.Error() + err2.Error() + err3.Error())
		}
		conn.Close()

		muuid := "M" + com.CreateUUID()
		d := []byte(muuid)
		d = append(d, 0xa)
		_, err1 = conn.Write(d)

		d = make([]byte, 2000)

		// receive b and c
		var getB bool = false
		var len, step int

		for i := 0; i < 15; i++ {
			for { //b
				conn.SetReadDeadline(time.Now().Add(time.Microsecond * 500)) //500ms
				_, _, err := conn.ReadFromUDP(d)
				if err != nil { //timeout
					break
				} else if err1 == nil && string(d[:37]) == muuid && d[37] == 0xb { //get b
					getB = true
					break
				}
			}
			for { //c
				conn.SetReadDeadline(time.Now().Add(time.Microsecond * 100)) //100ms
				_, _, err := conn.ReadFromUDP(d)
				if err != nil { //timeout, maybe sever offline
					err = errors.New("sever not has reply")
					return 0, err
				} else if err1 == nil && string(d[:37]) == muuid && d[37] == 0xc { //get c
					len = int(d[38])<<8 + int(d[39])
					step = int(d[40])<<8 + int(d[41])
					break
				}
			}
			if step == 1 {
				if getB {
					return uint16(len), nil
				}
				return uint16(len - 1), nil
			}

			step = step / 2
			d = []byte(muuid)
			// send d or e
			if getB { //e
				d = append(d, 0xe, uint8(len>>8), uint8(len), uint8(step>>8), uint8(step))
				_, err := conn.Write(d)
				if err != nil {
					return 0, err
				}
			} else {
				d = append(d, 0xd, uint8(len>>8), uint8(len), uint8(step>>8), uint8(step))
				_, err := conn.Write(d)
				if err != nil {
					return 0, err
				}
			}
		}

	}
	err := errors.New("Exception")
	return 0, err
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

	for {
		d := make([]byte, 2000)
		n, raddr, _ := handle.ReadFromUDP(d)
		var bodyB, bodyC []byte = d[:37], d[:37]
		bodyB = append(bodyB, 0xb)

		if n == 38 && d[37] == 0xa { //get a
			tmp := make([]byte, 1000)
			bodyB = append(bodyB, tmp...)
			bodyC = append(bodyC, 0xc, 3, 232, 3, 232) //len,step=1000

		} else if n == 42 && d[37] == 0xd { // get d
			len := int(d[38])<<8 + int(d[39])
			step := int(d[40])<<8 + int(d[41])
			len = len - step
			tmp := make([]byte, len)
			bodyB = append(bodyB, tmp...)
			bodyC = append(bodyC, 0xd, uint8(len>>8), uint8(len), d[40], d[41])

		} else if n == 42 && d[37] == 0xe { //get e
			len := int(d[38])<<8 + int(d[39])
			step := int(d[40])<<8 + int(d[41])
			len = len + step
			tmp := make([]byte, len)
			bodyB = append(bodyB, tmp...)
			bodyC = append(bodyC, 0xe, uint8(len>>8), uint8(len), d[40], d[41])
		}
		_, err := handle.WriteToUDP(bodyC, raddr) //reply c
		if err != nil {
			// log error
		}
		err = rawnet.SendDFIPPacket(lIP, raddr.IP, port, uint16(raddr.Port), bodyB) //reply b
		if err != nil {
			// log error
		}
	}
}
