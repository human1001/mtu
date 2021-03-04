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

	fmt.Println("_________:")
	fmt.Println(mtu.Client(true, true))
	// fmt.Println(mtu.Sever())
}
