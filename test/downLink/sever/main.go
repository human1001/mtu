package main

import (
	"fmt"
	"os"

	"github.com/lysShub/e"
	"github.com/lysShub/mtu"
)

func main() {

	fh, err := os.OpenFile("/root/mtuDiscover.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	e.L(os.Stderr, fh) // 设置错误日志

	// sever的公网IP为severIP
	m := mtu.NewMTU(func(m *mtu.MTU) *mtu.MTU {
		return m
	})

	fmt.Println("开始:")
	fmt.Println(m.Sever())
}
