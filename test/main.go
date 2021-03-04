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
	// 啊啊

	fmt.Println("_________:")
	fmt.Println(mtu.Client(true, true))
	// fmt.Println(mtu.Sever())
}
