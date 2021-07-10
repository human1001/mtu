package main

import (
	"fmt"

	"github.com/lysShub/mtu"
)

func main() {
	m := mtu.NewMTU(func(m *mtu.MTU) *mtu.MTU {
		m.SeverAddr = "114.116.254.26"
		return m
	})

	fmt.Println("开始:")
	fmt.Println(m.Client(false, false))
}
