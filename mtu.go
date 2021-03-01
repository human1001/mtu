package mtu

import (
	"bytes"
	"errors"
	"fmt"
	"mtu/internal/com"
	"mtu/internal/rawnet"
	"net"
	"os/exec"
	"runtime"
	cecf "seemfast/configs/client"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/ipv4"
)

//

// sever addr
const sever string = "" //ip or domain
// port port used by the server and client
const port uint16 = 19989

// MTU discovery
// 此处的MTU指UDP数据最大的长度(8+20+14=42), 一般数据链路层传输数据包大小

// Client client 阻塞一段时间
func Client(isUpLink bool) uint16 {

	if isUpLink { //上行链路 ping
		var d []byte
		var li [][]byte
		var n int
		var cmd *exec.Cmd
		var stdout, stderr bytes.Buffer
		var wrap []byte = make([]byte, 0)
		var F func(l int) int // 0:发生错误或异常,如超时等 -1:太小  1:太大

		// android darwin linux windows freebsd(未被支持)
		if runtime.GOOS == "windows" {
			wrap = append(wrap, []byte{13, 10, 13, 10}...) // windows换行：\r\n=回车换行; 用于分割空行
			F = func(l int) int {
				cmd = exec.Command("cmd", "/C", "ping", "-f", "-n", "1", "-l", strconv.Itoa(l), "-w", "1000", "baidu.com")
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
						if strings.Contains(rv, `TTL`) { //太小
							return -1
						} else if strings.Contains(rv, `DF`) {
							return 1
						}
					}
				}
				return 0
			}
		} else if runtime.GOOS == "linux" {
			wrap = append(wrap, []byte{10, 10}...) // Linux换行：\n=换行
			F = func(l int) int {
				// ping -M do -c 1 -s 1370 -W 1 baidu.com
				cmd = exec.Command("ping", "-M", "do", "-c", "1", "-s", strconv.Itoa(l), "-W", "1", "baidu.com")
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr
				err := cmd.Run()
				d = make([]byte, 0)
				if err != nil {
					d = append(d, []byte(err.Error())...)
				}
				d = append(d, stderr.Bytes()...)
				d = append(d, stdout.Bytes()...)
				d = com.ToUtf8(d[:n]) //编码判断

				li = bytes.Split(d, wrap)
				for _, v := range li {
					rv := string(v)
					if strings.Contains(rv, strconv.Itoa(l)) {
						if strings.Contains(rv, `ttl`) { //太小
							return -1
						} else if strings.Contains(rv, `too long`) {
							return 1
						}
					}
				}
				return 0
			}
		} else if runtime.GOOS == "darwin" {
			wrap = append(wrap, []byte{13, 13}...) // macOS换行：\r=回车
			F = func(l int) int {
				// ping -c 1 -t 1  -D -s 1500 baidu.com
				cmd = exec.Command("cmd", "/C", "ping", "-c", "1", "-t", "1", "-D", "-s", strconv.Itoa(l), "baidu.com")
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
						if strings.Contains(rv, `ttl`) { //太小
							return -1
						} else if strings.Contains(rv, `too long`) {
							return 1
						}
					}
				}
				return 0
			}
		} else if runtime.GOOS == "android" {
			wrap = append(wrap, []byte{10, 10}...) // Linux换行：\n=换行
			F = func(l int) int {
				cmd = exec.Command("/bin/ping", "-M", "do", "-c", "1", "-s", strconv.Itoa(l), "-W", "1", "baidu.com") //"-c", "1", "baidu.com"
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
						if strings.Contains(rv, `ttl`) { //太小
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
			com.Errorlog(err)
			return 0
		}

		// 二分法查找
		left, right, mid, step := 1, 2000, 0, 1999
		for {
			// mid向下取整
			mid = int(float64((left + right) / 2))
			r := F(mid)
			fmt.Println(mid, r, right-left)

			if 1 == r { //太大
				right = mid - 1
			} else if -1 == r { //太小
				left = mid + 1
			} else { // 获取错误
				break
			}
			step = right - left
			if step <= 3 {
				for i := right + 1; i <= left; i-- {
					if F(i) == -1 {
						return uint16(i)
					}
				}
			}
		}
	} else { //下行链路 precision = 10
		raddr1, err1 := net.ResolveUDPAddr("udp", cecf.SEVERADDR+":19989")
		laddr1, err2 := net.ResolveUDPAddr("udp", ":19989")
		conn, err3 := net.DialUDP("udp", laddr1, raddr1)
		if com.Errorlog(err1, err2, err3) {
			return 0
		}
		conn.Close()

		muuid := "M" + com.CreateUUID()
		d := []byte(muuid)
		d = append(d, 0xa)
		_, err1 = conn.Write(d)

		// 接收 b
		d = make([]byte, 2000)
		var mtu uint16 = 0
		go func() {
			for {
				n, _, err1 := conn.ReadFromUDP(d)
				if err1 == nil && string(d[:37]) == muuid && d[37] == 0xb {
					mtu = uint16(n)
				}
			}
		}()
		time.Sleep(time.Second)
		if mtu == 0 { //没有接收到数据包 服务器关闭
			err := errors.New("MTU discover sever maybe closed")
			com.Errorlog(err)
		} else {
			return mtu
		}
	}
	return 0
}

// Sever sever
func Sever() {
	laddr, err1 := net.ResolveUDPAddr("udp", ":19989")
	lh, err2 := net.ListenUDP("udp", laddr)
	if com.Errorlog(err1, err2) {
		return
	}
	defer lh.Close()

	lIP := rawnet.GetLocalIP()
	device := rawnet.GetDevice(lIP.String())
	if lIP == nil || device == "" {
		err1 = errors.New("can't get local IP or network card name；Loacl IP:" + lIP.String())
		com.Errorlog(err1)
		return
	}

	d := make([]byte, 2000)
	for {
		// fmt.Println("读取UDP数据")
		n, raddr, err1 := lh.ReadFromUDP(d)
		if com.Errorlog(err1) {
			continue
		}

		muuid := d[:37]
		if n == 38 {
			if d[37] == 0xa { //开始 探测下行链路
				// fmt.Println("读取到UDP数据", string(muuid), d[37])
				// 发送DF IP包
				go func() {
					ipAddr, err1 := net.ResolveIPAddr("ip4", lIP.String())
					conn, err2 := net.ListenIP("ip4:udp", ipAddr)
					if com.Errorlog(err1, err2) {
						return
					}
					defer conn.Close()

					rawConn, err1 := ipv4.NewRawConn(conn) //ipconn to rawconn
					if com.Errorlog(err1) {
						return
					}
					var err error
					for i := 0; i < 36; i++ {
						var dr []byte
						if i < 30 {
							dr = make([]byte, 1500-i*10) //1250-(8+20+14)
						} else {
							dr = make([]byte, 560-(i-30)*10)
						}
						for i := 0; i < 37; i++ {
							dr[i] = muuid[i]
						}
						dr[37] = 0xb

						uR := rawnet.PackageUDP(lIP, raddr.IP, 19989, uint16(raddr.Port), dr)
						iph, err1 := rawnet.PackageIPHeader(lIP, raddr.IP, 19989, uint16(raddr.Port), 2, 0, 17, uR)
						if com.Errorlog(err1) {
							return
						}

						err = rawConn.WriteTo(iph, uR, nil)
					}
					com.Errorlog(err)
					return
				}()
			}
		}
	}
}
