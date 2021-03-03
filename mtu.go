package mtu

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mtu/internal/com"
	"mtu/internal/rawnet"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

// Discover the MTU of the link by UDP packet

// sever sever addr, ip or domain
const sever string = "114.116.254.26"

// port port used by the server and client
const port uint16 = 19989

// pingHost ping host
const pingHost string = "baidu.com"

// UpLinkFast use PMTUD get MTU, more fast less reliable
// eg: ping: local error: Message too long, mtu=1400
const UpLinkFast bool = true

// Client client
// if isUpLink = false, it will discover downlink's mtu, need sever support
// discover the uplink through the PING command
// may block for ten seconds; for example, PING command didn't replay
func Client(isUpLink bool) (uint16, error) {

	if isUpLink { //Uplink ping

		var cmd *exec.Cmd
		var Out, Err io.ReadCloser
		var stdout, stderr []byte
		var err1, err2 error
		var wrap []byte = make([]byte, 0)
		var F func(l int) int // 0:error or exception,eg:timeout; -1:too small  1:too big

		var Fs func(f bool, b []byte) int = func(f bool, b []byte) int {
			if UpLinkFast {
				if bytes.Contains(stderr, []byte("mtu=")) { // Linux Wrap：\n 10
					a := bytes.Split(stderr, []byte("mtu="))[1]
					for i, v := range a {
						if v == uint8(10) {
							j, err := strconv.Atoi(string(a[:i]))
							if err != nil {
								break
							}
							return j - 28
						}
					}
				}
			}
			return 1 //too long
		}

		if runtime.GOOS == "windows" {
			//  Used to split blank lines
			wrap = append(wrap, []byte{13, 10, 13, 10}...) // windows Wrap：\r\n=Enter(13) Wrap(10)
			F = func(l int) int {
				cmd = exec.Command("cmd", "/C", "ping", "-f", "-n", "1", "-l", strconv.Itoa(l), "-w", "1000", pingHost)
				Out, err1 = cmd.StdoutPipe()
				Err, err2 = cmd.StderrPipe()
				if err1 != nil || err2 != nil {
					return 0
				}
				cmd.Start()
				stdout, err1 = ioutil.ReadAll(Out)
				stderr, err2 = ioutil.ReadAll(Err)
				if err1 != nil || err2 != nil {
					return 0
				}
				// cmd.Wait()
				stdout = com.ToUtf8(stdout)

				if bytes.Contains(stdout, []byte("DF")) {
					return 1 //too long
				} else if bytes.Contains(stdout, []byte("ms")) && bytes.Contains(stdout, []byte(strconv.Itoa(l))) {
					return -1 //too small
				} else {
					return 0
				}
			}
		} else if runtime.GOOS == "linux" {
			wrap = append(wrap, []byte{10, 10}...) // Linux Wrap：\n=Wrap
			F = func(l int) int {
				cmd = exec.Command("ping", "-M", "do", "-c", "1", "-s", strconv.Itoa(l), "-W", "1", pingHost)
				Out, err1 = cmd.StdoutPipe()
				Err, err2 = cmd.StderrPipe()
				if err1 != nil || err2 != nil {
					return 0
				}
				cmd.Start()
				stdout, err1 = ioutil.ReadAll(Out)
				stderr, err2 = ioutil.ReadAll(Err)
				if err1 != nil || err2 != nil {
					return 0
				}
				//cmd.Wait()
				stdout = com.ToUtf8(stdout)
				stderr = com.ToUtf8(stderr)

				if bytes.Contains(stderr, []byte("too long")) {
					// ping: local error: message too long, mtu=1400
					return Fs(UpLinkFast, stderr)
				} else if bytes.Contains(stdout, []byte("ms")) && bytes.Contains(stdout, []byte(strconv.Itoa(l))) {
					return -1 //too small
				} else {
					return 0
				}
			}
		} else if runtime.GOOS == "darwin" {
			wrap = append(wrap, []byte{13, 13}...) // macOS Wrap：\r=Enter
			F = func(l int) int {
				cmd = exec.Command("cmd", "/C", "ping", "-c", "1", "-t", "1", "-D", "-s", strconv.Itoa(l), pingHost)
				Out, err1 = cmd.StdoutPipe()
				Err, err2 = cmd.StderrPipe()
				if err1 != nil || err2 != nil {
					return 0
				}
				cmd.Start()
				stdout, err1 = ioutil.ReadAll(Out)
				stderr, err2 = ioutil.ReadAll(Err)
				if err1 != nil || err2 != nil {
					return 0
				}
				//cmd.Wait()
				stdout = com.ToUtf8(stdout)
				stderr = com.ToUtf8(stderr)

				if bytes.Contains(stderr, []byte("too long")) {
					return 1

				} else if bytes.Contains(stdout, []byte("ms")) && bytes.Contains(stdout, []byte(strconv.Itoa(l))) {
					// PING baidu.com (39.156.69.79) 1000(1028) bytes of data.
					// 1008 bytes from 39.156.69.79: icmp_seq=1 ttl=51 time=85.3 ms
					return -1 //too small
				} else {
					return 0
				}
			}
		} else if runtime.GOOS == "android" {

			wrap = append(wrap, []byte{10, 10}...) // Linux Wrap：\n=Wrap
			F = func(l int) int {
				cmd = exec.Command("/bin/ping", "-M", "do", "-c", "1", "-s", strconv.Itoa(l), "-W", "1", pingHost) //"-c", "1", pingHost
				Out, err1 = cmd.StdoutPipe()
				Err, err2 = cmd.StderrPipe()
				if err1 != nil || err2 != nil {
					return 0
				}
				cmd.Start()
				stdout, err1 = ioutil.ReadAll(Out)
				stderr, err2 = ioutil.ReadAll(Err)
				if err1 != nil || err2 != nil {
					return 0
				}
				//cmd.Wait()
				stdout = com.ToUtf8(stdout)
				stderr = com.ToUtf8(stderr)

				if bytes.Contains(stderr, []byte("too long")) {
					// ping: local error: Message too long, mtu=1400
					return Fs(UpLinkFast, stderr)
				} else if bytes.Contains(stdout, []byte("ms")) && bytes.Contains(stdout, []byte(strconv.Itoa(l))) {
					// PING baidu.com (39.156.69.79) 1000(1028) bytes of data.
					// 1008 bytes from 39.156.69.79: icmp_seq=1 ttl=51 time=85.3 ms
					return -1 //too small
				} else {
					return 0
				}
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
			fmt.Println(mid, left, right)

			if 1 == r { //big
				right = mid - 1
			} else if -1 == r { //small
				left = mid + 1
			} else if 0 == r { // r==0 error or exception
				break
			} else {
				return uint16(r), nil
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
		defer conn.Close()

		muuid := "M" + com.CreateUUID()
		d := []byte(muuid)
		d = append(d, 0xa)
		_, err1 = conn.Write(d)

		d = make([]byte, 2000)
		// receive b and c
		var getB, getC bool = false, false
		var len, step int

		for i := 0; i < 15; i++ {

			getB, getC = false, false
			for {
				conn.SetReadDeadline(time.Now().Add(time.Second))
				_, _, err := conn.ReadFromUDP(d)
				if err != nil && !(getB || getC) {
					return 0, errors.New("sever not has reply")

				} else if err != nil && getC { // too long
					break
				} else if err1 == nil && string(d[:37]) == muuid && d[37] == 0xc { //get c
					len = int(d[38])<<8 + int(d[39])
					step = int(d[40])<<8 + int(d[41])
					getC = true
					if getB { // get b and get c
						break
					}
				} else if err1 == nil && string(d[:37]) == muuid && d[37] == 0xb { //get b
					getB = true
					if getC { // get b and get c
						break
					}
				}
				fmt.Println("读取到")
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
		var bodyB, bodyC []byte = make([]byte, 37), make([]byte, 37)
		copy(bodyB, d[:37])
		copy(bodyC, d[:37])
		bodyB = append(bodyB, 0xb)

		fmt.Println(d[:n])

		if n == 38 && d[37] == 0xa { //get a
			fmt.Println("收到a")

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
		err = rawnet.SendIPPacketDFUDP(lIP, raddr.IP, port, uint16(raddr.Port), bodyB) //reply b
		if err != nil {
			// log error
		}
	}
}
