package main

import (
	"fmt"

	"github.com/lysShub/mtu"
)

func main() {

	m := mtu.NewMTU(func(m *mtu.MTU) *mtu.MTU {
		m.PingHost = "baidu.com"
		return m
	})

	fmt.Println("开始:")
	fmt.Println(m.Client(true, false))
	// fmt.Println(m.Sever())
}
