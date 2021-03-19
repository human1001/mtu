package main

import (
	"fmt"
	"io"
	"os"

	"github.com/lysShub/mtu"
	"github.com/lysShub/mtu/internal/com"
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
