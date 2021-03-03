// +build windows

package com

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
)

// subPingDF don's support fast mode
func subPingDF(l int, pingHost string, faster bool) (int, error) {

	cmd := exec.Command("cmd", "/C", "ping", "-f", "-n", "1", "-l", strconv.Itoa(l), "-w", "1000", pingHost)

	Out, err1 := cmd.StdoutPipe()
	if err1 != nil {
		return 0, err1
	}
	cmd.Start()
	stdout, err1 := ioutil.ReadAll(Out)

	if err1 != nil {
		return 0, err1
	}
	// cmd.Wait()
	stdout = ToUtf8(stdout)

	if bytes.Contains(stdout, []byte("DF")) {
		return 1, nil //too long
	} else if bytes.Contains(stdout, []byte("ms")) && bytes.Contains(stdout, []byte(strconv.Itoa(l))) {
		return -1, nil //too small
	} else {
		fmt.Println("return 0")
		return 0, nil
	}

}
