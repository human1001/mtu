package main

import (
	"errors"
	"fmt"

	"github.com/lysShub/mtu"
	"github.com/lysShub/mtu/internal/com"
)

func main() {

	var err = errors.New("aa")
	var err1 = errors.New("afasdfasa")
	com.Errlog(err, err1)
	return

	m := mtu.NewMTU(func(m *mtu.MTU) *mtu.MTU {
		m.PingHost = "baidu.com"
		return m
	})

	fmt.Println("开始:")
	fmt.Println(m.Client(true, true))
}
