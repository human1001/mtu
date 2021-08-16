package mtu

import (
	"errors"
	"net"
	"strings"
	"time"

	"github.com/lysShub/mtu/internal/com"
	"github.com/lysShub/mtu/internal/ping"
	"github.com/lysShub/mtu/internal/rawnet"
	"github.com/lysShub/tq"
)

// clientDownLink 下行链路MTU
func clientDownLink(sever string, port int) (uint16, error) {

	conn, err := net.DialUDP("udp", &net.UDPAddr{IP: nil, Port: port}, &net.UDPAddr{IP: net.ParseIP(sever), Port: port})
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	var mtu uint16
	var daCh chan []byte = make(chan []byte, 1)

	var da = []byte{3, 2000 >> 8, 2000 % (1 << 8), 1000 >> 8, 1000 % (1 << 8)} // 探测范围 [0,2000]
	daCh <- da
	var length, MaxDFlen, tmpStep int
	var step int = 1000 // 和da初始值对应

	var end bool = false

	// 读
	go func() {
		var l int
		for !end {
			da = make([]byte, 2000)
			if l, err = conn.Read(da); err == nil && l >= 5 {

				if da[0] == 1 {
					if l > MaxDFlen {
						MaxDFlen = l
					}
				} else if da[0] == 2 {
					tmpStep = int(da[3])<<8 + int(da[4])

					if tmpStep <= step {
						length, step = int(da[1])<<8+int(da[2]), int(da[3])<<8+int(da[4])

						if step <= 1 {
							if length <= MaxDFlen { // 能收到
								mtu = uint16(length)
							} else {
								mtu = uint16(length) - 1
							}
							end = true
						}
						step = step>>1 + step&0b1
						if length <= MaxDFlen {
							daCh <- []byte{4, byte(length >> 8), byte(length), byte(step >> 8), byte(step)}
						} else {
							daCh <- []byte{3, byte(length >> 8), byte(length), byte(step >> 8), byte(step)}
						}
					}
				}
			}
		}
	}()

	var data []byte = make([]byte, 0)
	for i := 0; i < 40 && !end; i++ { // 超时2s
		if len(daCh) == 0 {
			if _, err = conn.Write(data); err != nil {
				return 0, err
			}
			time.Sleep(time.Millisecond * 20) // 50
		} else {
			data = <-daCh
			i = 0
		}
	}
	if mtu == 0 {
		return 0, errors.New("sever timeout")
	} else {
		return mtu + 28, nil // 8+20
	}
}

// clientUpLink 上行链路MTU
func clientUpLink(pingHost string, faster bool) (uint16, error) {

	if faster { // 快速模式，猜测常见大小
		var f int
		for i := 1472; i <= 1473; i++ {
			r, err := ping.PingDF(i, pingHost, true)
			if err != nil {
				return 0, err
			} else if err == nil && r > 1 {
				return uint16(r) + 28, nil
			}
			f += r // 1372: -1, 1473:1
		}
		if f == 0 {
			return 1472 + 28, nil
		}
		f = 0
		for i := 1372; i <= 1373; i++ {
			r, err := ping.PingDF(i, pingHost, false)
			if err != nil {
				return 0, err
			} else if err == nil && r > 1 {
				return uint16(r) + 28, nil
			}
			f += r
		}
		if f == 0 {
			return 1372 + 28, nil
		}
	}

	// 二分法
	left, right, mid, step := 1, 2000, 0, 1999
	for {
		mid = int(float64((left + right) / 2))
		r, err := ping.PingDF(mid, pingHost, faster)
		if err != nil {
			return 0, err
		}

		if r == 1 { //big
			right = mid - 1
		} else if -1 == r { //small
			left = mid + 1
		} else if r == 0 { // r==0 error or exception
			break
		} else {
			return uint16(r) + 28, nil
		}
		step = right - left
		if step <= 3 {
			for i := right + 1; i <= left; i-- {
				n, err := ping.PingDF(i, pingHost, faster)
				if n == -1 {
					return uint16(i) + 28, nil
				} else if err != nil {
					return 0, err
				}
			}
		}
	}
	return 0, nil
}

func (m *MTU) sever() error {

	var conn *net.UDPConn
	if conn, err = net.ListenUDP("udp", &net.UDPAddr{IP: nil, Port: m.Port}); err != nil {
		return err
	}
	defer conn.Close()

	Q := tq.NewTQ() // 时间任务队列

	var id int64
	var s map[int64]w = make(map[int64]w)

	var lIP net.IP
	if lIP = rawnet.GetLocalIP(); lIP == nil {
		return errors.New("can't get local IP")
	}

	var length, newLength, step int
	go func() {
		var da, stuff []byte = make([]byte, 64), make([]byte, 0)
		var raddr *net.UDPAddr
		var n, severMtu int = 0, 2000
		var delay time.Time
		for {
			if n, raddr, err = conn.ReadFromUDP(da); !com.Errlog(err) && n >= 5 {

				length, step = int(da[1])<<8+int(da[2]), int(da[3])<<8+int(da[4])

				if length-step >= 1 {
					if da[0] == 3 { // 减
						newLength = length - step
					} else if da[0] == 4 { // 加
						newLength = length + step
					}
					stuff = make([]byte, newLength-1)

					if newLength < severMtu {
						err = rawnet.SendIPPacketDFUDP(lIP, raddr.IP, uint16(m.Port), uint16(raddr.Port), append([]byte{1}, stuff...))
						if err != nil {
							if strings.Contains(err.Error(), "message too long") {
								severMtu = newLength
							} else {
								com.Errlog(err)
							}
						} else {
							delay = time.Now().Add(time.Millisecond * 30)
						}
					} else {
						delay = time.Now().Add(time.Millisecond * 0)
					}

					var t w
					t.data = []byte{2, byte(newLength >> 8), byte(newLength), byte(step >> 8), byte(step)}
					t.raddr = *raddr
					s[id] = t
					Q.Add(tq.Ts{
						T: delay,
						P: id,
					})
					id++ // 不超过int64容量
				}
			}
		}
	}()

	// 发送 2
	var r interface{}
	for {
		r = <-(Q.MQ)
		v, ok := r.(int64)
		if ok {
			raddr := s[v].raddr

			_, err = conn.WriteToUDP(s[v].data, &raddr)
			com.Errlog(err)
		}
	}
}

type w struct {
	raddr net.UDPAddr
	data  []byte
}
