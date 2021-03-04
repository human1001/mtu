// +build linux android netbsd openbsd freebsd

package com

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os/exec"
	"strconv"
)

// subPingDF linux
func subPingDF(l int, pingHost string, faster bool) (int, error) {

	cmd := exec.Command("/bin/ping", "-M", "do", "-c", "1", "-s", strconv.Itoa(l), "-w", "1", pingHost)

	Out, err1 := cmd.StdoutPipe()
	Err, err2 := cmd.StderrPipe()
	if err1 != nil || err2 != nil {
		return 0, errors.New(err1.Error() + err2.Error())
	}
	cmd.Start()
	stdout, err1 := ioutil.ReadAll(Out)
	stderr, err2 := ioutil.ReadAll(Err)
	Out.Close()

	if err1 != nil || err2 != nil {
		return 0, errors.New(err1.Error() + err2.Error())
	}

	//cmd.Wait()

	stdout = ToUtf8(stdout)
	stderr = ToUtf8(stderr)

	if bytes.Contains(stderr, []byte("too long")) {
		// ping: local error: message too long, mtu=1400
		return fast(faster, stderr), nil

	} else if bytes.Contains(stdout, []byte("ms")) && bytes.Contains(stdout, []byte(strconv.Itoa(l))) {
		return -1, nil //too small
	} else {
		return 0, nil
	}
}
