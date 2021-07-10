// +build windows

package ping

import (
	"bytes"
	"errors"
	"os/exec"
	"strconv"

	"github.com/lysShub/mtu/internal/com"
)

// subPingDF
func subPingDF(l int, pingHost string, faster bool) (int, error) {

	// err 对应为stdErr输出, 类似为 exit status 1; 没有参考价值、直接忽略
	// ping -f -n 1 -l 1000 -w 1000 baidu.com
	da, _ := exec.Command("cmd", "/C", "ping", "-f", "-n", "1", "-l", strconv.Itoa(l), "-w", "1000", pingHost).CombinedOutput()

	da = com.ToUtf8(da)

	if bytes.Contains(da, []byte("DF")) {
		return 1, nil //too long
	} else if bytes.Contains(da, []byte("ms")) && bytes.Contains(da, []byte(strconv.Itoa(l))) {
		return -1, nil //too small
	} else {
		return 0, errors.New("ping command output: " + Wrap() + string(com.ToUtf8(da)) + Wrap() + "is not normal")
	}
}

func wrap() string {
	return `\r\n`
}
