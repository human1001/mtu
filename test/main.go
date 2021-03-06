package main

import (
	"fmt"
	"io"
	"mtu"
	"mtu/internal/com"
	"os"
)

func main() {
	com.Writers = []io.Writer{
		os.Stdout,
	}

	mtu.PingHost = "baidu.com"
	mtu.Port = uint16(19986)

	fmt.Println("_________:")
	fmt.Println(mtu.Client(true, true))
	// fmt.Println(m.Sever())
}
