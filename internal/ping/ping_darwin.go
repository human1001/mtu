// +build darwin

package ping

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os/exec"
	"strconv"
)

// subPingDF
func subPingDF(l int, pingHost string, faster bool) (int, error) {

	da, _ := exec.Command("ping", "-D", "-c", "1", "-s", strconv.Itoa(l), "-t", "1", pingHost).CombinedOutput()

	stdout = com.ToUtf8(da)

	if bytes.Contains(da, []byte("too long")) {
		return 1, nil

	} else if bytes.Contains(da, []byte("ms")) && bytes.Contains(da, []byte(strconv.Itoa(l))) {
		return -1, nil //too small
	} else if bytes.Contains(da, []byte("100.0%")) && bytes.Contains(da, []byte(strconv.Itoa(l))) {
		// sometimes, too long will don't echo "too long" instead timeout
		return 1, nil //too long
	} else {
		return 0, errors.New("ping command output: " + Wrap() + string(da) + Wrap() + "is not normal")
	}

	// Out, err1 := cmd.StdoutPipe()
	// Err, err2 := cmd.StderrPipe()
	// if err1 != nil || err2 != nil {
	// 	return 0, errors.New(err1.Error() + err2.Error())
	// }
	// cmd.Start()
	// stdout, err1 := ioutil.ReadAll(Out)
	// stderr, err2 := ioutil.ReadAll(Err)
	// if err1 != nil || err2 != nil {
	// 	return 0, errors.New(err1.Error() + err2.Error())
	// }
	// //cmd.Wait()
	// stdout = ToUtf8(stdout)
	// stderr = ToUtf8(stderr)

	// if bytes.Contains(stderr, []byte("too long")) {
	// 	return 1, nil

	// } else if bytes.Contains(stdout, []byte("ms")) && bytes.Contains(stdout, []byte(strconv.Itoa(l))) {
	// 	return -1, nil //too small
	// } else if bytes.Contains(stdout, []byte("100.0%")) && bytes.Contains(stdout, []byte(strconv.Itoa(l))) {
	// 	// sometimes, too long will don't echo "too long" instead timeout
	// 	return 1, nil //too long
	// } else {
	// 	return 0, errors.New("ping command output: " + Wrap() + string(da) + Wrap() + "is not normal")
	// }
}

func wrap() string {
	return `\n`
}
