//go:build linux || android || netbsd || openbsd || freebsd
// +build linux android netbsd openbsd freebsd

package ping

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/human1001/mtu/internal/com"
)

// subPingDF linux
func subPingDF(l int, pingHost string, faster bool) (int, error) {

	// /bin/ping -M do -c 1 -s 1500 -w 1 baidu.com
	da, _ := exec.Command("/bin/ping", "-M", "do", "-c", "1", "-s", strconv.Itoa(l), "-w", "1", pingHost).CombinedOutput()

	da = com.ToUtf8(da)

	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println(string(da))
	fmt.Println()
	fmt.Println()
	fmt.Println()

	if bytes.Contains(da, []byte("too long")) {
		// ping: local error: message too long, mtu=1400
		return fastMode(faster, da), nil

	} else if bytes.Contains(da, []byte("ms")) && bytes.Contains(da, []byte(strconv.Itoa(l))) {
		return -1, nil //too small
	} else {
		return 0, errors.New("ping command output: " + Wrap() + string(da) + Wrap() + "is not normal")
	}
}

func fastMode(fast bool, stderr []byte) int {
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

func wrap() string {
	return `\r`
}
