// +build darwin

package com

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os/exec"
	"strconv"
)

// subPingDF darwin
func subPingDF(l int, pingHost string, faster bool) (int, error) {

	cmd := exec.Command("cmd", "/C", "ping", "-c", "1", "-t", "1", "-D", "-s", strconv.Itoa(l), pingHost)
	Out, err1 := cmd.StdoutPipe()
	Err, err2 := cmd.StderrPipe()
	if err1 != nil || err2 != nil {
		return 0, errors.New(err1.Error() + err2.Error())
	}
	cmd.Start()
	stdout, err1 := ioutil.ReadAll(Out)
	stderr, err2 := ioutil.ReadAll(Err)
	if err1 != nil || err2 != nil {
		return 0, errors.New(err1.Error() + err2.Error())
	}
	//cmd.Wait()
	stdout = ToUtf8(stdout)
	stderr = ToUtf8(stderr)

	if bytes.Contains(stderr, []byte("too long")) {
		return 1, nil

	} else if bytes.Contains(stdout, []byte("ms")) && bytes.Contains(stdout, []byte(strconv.Itoa(l))) {
		// PING baidu.com (39.156.69.79) 1000(1028) bytes of data.
		// 1008 bytes from 39.156.69.79: icmp_seq=1 ttl=51 time=85.3 ms
		return -1, nil //too small
	} else {
		return 0, nil
	}
}
