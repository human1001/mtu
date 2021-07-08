// +build windows

package rawnet

func sendUDP() {
	var wsaData windows.WSAData
	if err = windows.WSAStartup(514, &wsaData); err != nil {
		fmt.Println(err)
		return
	}
	var fd windows.Handle
	if fd, err = windows.Socket(windows.AF_INET, windows.SOCK_DGRAM, windows.IPPROTO_UDP); err != nil {
		return
	}

	//
	var rAddr windows.RawSockaddrInet4
	rAddr.Family = windows.AF_INET
	rAddr.Port = 19986
	rAddr.Addr = [4]byte{172, 26, 167, 96}
	var sAddr windows.Sockaddr
	if sAddr, err = (*windows.RawSockaddrAny)(unsafe.Pointer(&rAddr)).Sockaddr(); err != nil {
		fmt.Println(err)
		return
	}
	var da []byte = make([]byte, 1370)
	var co int = 0
	go func() {
		for {
			time.Sleep(time.Second)
			fmt.Println(co>>20, "MB/s")
			co = 0
		}
	}()

	for {
		if err = windows.Sendto(fd, da, 0, sAddr); err != nil {
			fmt.Println(err)
			return
		}
		co += 1372
		// time.Sleep(time.Second)
	}

}
