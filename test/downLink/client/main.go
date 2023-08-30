package main

import (
	"fmt"

	"github.com/human1001/mtu"
)

func main() {
	m := mtu.NewMTU(func(m *mtu.MTU) *mtu.MTU {
		m.SeverAddr = "127.0.0.1"
		return m
	})

	fmt.Println("开始:")
	fmt.Println(m.Client(false, false))
}
