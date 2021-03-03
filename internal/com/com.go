package com

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gogs/chardet"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/net/html/charset"
)

// ToUtf8 Convert to any encoding (as far as possible) to utf8 encoding
func ToUtf8(s []byte) []byte {

	// chardet echo charsets:Shift_JIS,EUC-JP,EUC-KR,Big5,GB18030,ISO-8859-2(windows-1250),ISO-8859-5,ISO-8859-6,ISO-8859-7,indows-1253,ISO-8859-8(windows-1255),ISO-8859-8-I,ISO-8859-9(windows-1254),windows-1256,windows-1251,KOI8-R,IBM424_rtl,IBM424_ltr,IBM420_rtl,IBM420_ltr,ISO-2022-JP

	d := chardet.NewTextDetector() //chardet is more precise charset.DetermineEncoding
	var rs *chardet.Result
	var err1 error
	if len(s) > 1024 {
		if utf8.Valid(s[:1024]) {
			return s
		}
		rs, err1 = d.DetectBest(s[:1024])
	} else {
		if utf8.Valid(s) {
			return s
		}
		rs, err1 = d.DetectBest(s)
	}

	// charset input charsets:utf-8,ibm866,iso-8859-2,iso-8859-3,iso-8859-4,iso-8859-5,iso-8859-6,iso-8859-7,iso-8859-8,iso-8859-8-i,iso-8859-10,iso-8859-13,iso-8859-14,iso-8859-15,iso-8859-16,koi8-r,koi8-u,macintosh,windows-874,windows-1250,windows-1251,windows-1252,windows-1253,windows-1254,windows-1255,windows-1256,windows-1257,windows-1258,x-mac-cyrillic,gbk,gb18030,big5,euc-jp,iso-2022-jp,shift_jis,euc-kr,replacement,utf-16be,utf-16le,x-user-defined,

	var maps map[string]string = make(map[string]string)
	maps = map[string]string{
		"Shift_JIS":    "shift_jis",
		"EUC-JP":       "euc-jp",
		"EUC-KR":       "euc-kr",
		"Big5":         "big5",
		"GB18030":      "gb18030",
		"ISO-8859-2 ":  "iso-8859-2",
		"ISO-8859-5":   "iso-8859-5",
		"ISO-8859-6":   "iso-8859-6",
		"ISO-8859-7":   "iso-8859-7",
		"ISO-8859-8":   "iso-8859-8",
		"ISO-8859-8-I": "iso-8859-8-i",
		"ISO-8859-9":   "iso-8859-10",
		"windows-1256": "windows-1256",
		"windows-1251": "windows-1251",
		"KOI8-R":       "koi8-r",
		"ISO-2022-JP":  "iso-2022-jp",
		"UTF-16BE ":    "utf-16be",
		"UTF-16LE ":    "utf-16le",
	}

	ct := maps[rs.Charset]
	if ct == "" || err1 != nil { // use charset.DetermineEncoding
		_, name, b := charset.DetermineEncoding([]byte(s), "utf-8")
		if b {
			return s
		}
		ct = name
	}

	byteReader := bytes.NewReader(s)
	reader, err1 := charset.NewReaderLabel(ct, byteReader)
	r, err2 := ioutil.ReadAll(reader)

	if err1 != nil || err2 != nil {
		return s
	}
	return r
}

// CreateUUID create id
// eg：312d6891-56d6-47ac-a266-b6bd56462d0e
func CreateUUID() string {
	u := uuid.Must(uuid.NewV4(), nil)
	return u.String()
}

// ClientDownLink client downlink MTU
func ClientDownLink(sever string, port uint16) (uint16, error) {
	raddr1, err1 := net.ResolveUDPAddr("udp", sever+":"+strconv.Itoa(int(port)))
	laddr1, err2 := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(int(port)))
	conn, err3 := net.DialUDP("udp", laddr1, raddr1)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, errors.New(err1.Error() + err2.Error() + err3.Error())
	}
	defer conn.Close()

	muuid := "M" + CreateUUID()
	d := []byte(muuid)
	d = append(d, 0xa)
	_, err1 = conn.Write(d)
	s := time.Now().UnixNano()

	var delay int64 = 1000
	d = make([]byte, 2000)
	// receive b and c
	var getB, getC bool = false, false
	var len, step int

	for i := 0; i < 15; i++ {

		getB, getC = false, false
		for {

			err := conn.SetReadDeadline(time.Now().Add((time.Millisecond * time.Duration(delay))))
			if err != nil {
				return 0, err
			}
			d = make([]byte, 2000)
			_, _, err = conn.ReadFromUDP(d)

			if s != 0 && string(d[:37]) == muuid {
				delay = (5 * (time.Now().UnixNano() - s) / 1e6) / 2
				fmt.Println("delay", delay)
				s = 0
			}

			if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
				return 0, err

			} else if errors.Is(err, os.ErrDeadlineExceeded) && !getC {
				return 0, errors.New("Server offline or Network fluctuations")

			} else if errors.Is(err, os.ErrDeadlineExceeded) && getC {
				break

			} else if err1 == nil && string(d[:37]) == muuid && d[37] == 0xc { //get c
				len = int(d[38])<<8 + int(d[39])
				step = int(d[40])<<8 + int(d[41])
				getC = true
				if getB {
					break
				}
			} else if err1 == nil && string(d[:37]) == muuid && d[37] == 0xb { //get b
				getB = true
				if getC {
					break
				}
			}
		}

		if step == 1 {
			if getB {
				return uint16(len), nil
			}
			return uint16(len - 1), nil
		}
		fmt.Println(step)
		step = step / 2
		d = []byte(muuid)
		if getB { //e

			d = append(d, 0xe, uint8(len>>8), uint8(len), uint8(step>>8), uint8(step))
			_, err := conn.Write(d)
			if err != nil {
				return 0, err
			}
		} else { // d

			d = append(d, 0xd, uint8(len>>8), uint8(len), uint8(step>>8), uint8(step))
			_, err := conn.Write(d)
			if err != nil {
				return 0, err
			}
		}
	}
	return 0, errors.New("Exception")
}

// fast uplink fast mode; faster and less reliable
func fast(fast bool, stderr []byte) int {
	if fast {
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
	return 1
}

func pingDF(l int, pingHost string, faster bool) (int, error) {

	return subPingDF(l, pingHost, faster)
}

// ClientUpLink client uplink
func ClientUpLink(pingHost string, faster bool) (uint16, error) {
	if faster {
		var f int
		for i := 1472; i <= 1473; i++ {
			r, err := pingDF(i, pingHost, false)
			if err != nil {
				return 0, err
			}
			f += r
		}
		if f == 0 {
			return 1472, nil
		}
		f = 0
		for i := 1372; i <= 1373; i++ {
			r, err := pingDF(i, pingHost, false)
			if err != nil {
				return 0, err
			}
			f += r
		}
		if f == 0 {
			return 1372, nil
		}
	}

	// Binary search
	left, right, mid, step := 1, 2000, 0, 1999
	for {
		mid = int(float64((left + right) / 2))
		r, err := pingDF(mid, pingHost, faster)
		if err != nil {
			return 0, err
		}

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
				n, err := pingDF(i, pingHost, faster)
				if n == -1 {
					return uint16(i), nil
				} else if err != nil {
					return 0, err
				}
			}
		}
	}
	return 0, nil
}
