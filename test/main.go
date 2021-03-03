package main

import (
	"fmt"
	"mtu"
)

func main() {
	fmt.Println("开始")
	fmt.Println(mtu.Client(true, true))
	// fmt.Println(mtu.Sever())
}
