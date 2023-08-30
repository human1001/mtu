package main

import (
	"fmt"

	"github.com/human1001/mtu"
)

func main() {

	// sever的公网IP为severIP
	m := mtu.NewMTU(func(m *mtu.MTU) *mtu.MTU {
		return m
	})

	fmt.Println("开始:")
	fmt.Println(m.Sever())
}
