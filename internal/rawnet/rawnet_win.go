// +build windows

package rawnet

import (
	"net"
	"unsafe"

	"golang.org/x/sys/windows"
)

//SendIPPacketDFWin windows send DF ip packet
func SendIPPacketDFWin(rIP net.IP, rPort, protocol uint16, d []byte) error {
	var wsaData windows.WSAData
	err := windows.WSAStartup(2<<16+2, &wsaData)
	if err != nil {
		return err
	}
	sh, err := windows.Socket(windows.AF_INET, windows.SOCK_RAW, int(protocol)) //windows.IPPROTO_ICMP
	if err != nil {
		return err
	}
	var rAddr windows.RawSockaddrInet4
	rAddr.Family = windows.AF_INET
	rAddr.Port = rPort
	ips := [4]byte{rIP[12], rIP[13], rIP[14], rIP[15]}
	rAddr.Addr = ips
	q := (*windows.RawSockaddrAny)(unsafe.Pointer(&rAddr))
	sAddr, err := q.Sockaddr()
	if err != nil {
		return err
	}
	var aOptVal bool = true
	err = windows.Setsockopt(sh, windows.IPPROTO_IP, 14, (*byte)(unsafe.Pointer(&aOptVal)), int32(unsafe.Sizeof(aOptVal)))
	if err != nil {
		return err
	}

	err = windows.Sendto(sh, d, 0, sAddr) //Complete transport layer data package
	if err != nil {
		return err
	}

	err = windows.Closesocket(sh)
	if err != nil {
		return err
	}
	err = windows.WSACleanup()
	if err != nil {
		return err
	}
	return nil
}

/*
* Corresponding C program
 */
// -------------------------
//  Windows
// -------------------------
// #ifndef UNICODE
// #define UNICODE
// #endif
// #define WIN32_LEAN_AND_MEAN
// #include <winsock2.h>
// #include <Ws2tcpip.h>
// #include <stdio.h>
// // Link with ws2_32.lib
// #pragma comment(lib, "Ws2_32.lib")
// int main()
// {
//     int iResult;
//     WSADATA wsaData;
//     SOCKET SendSocket = INVALID_SOCKET;
//     // sockaddr_in RecvAddr;
//     SOCKADDR_IN RecvAddr;
//     unsigned short Port = 27015;
//     char SendBuf[8] = {8,0,247,95,0,1,0,159};
//     int BufLen = 8;
//     //----------------------
//     // Initialize Winsock
//     iResult = WSAStartup(MAKEWORD(2, 2), &wsaData);
//     if (iResult != NO_ERROR)
//     {
//         wprintf(L"WSAStartup failed with error: %d\n", iResult);
//         return 1;
//     }
//     //---------------------------------------------
//     // Create a socket for sending data
//     SendSocket = socket(AF_INET, SOCK_RAW, IPPROTO_ICMP);
//     // SendSocket = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
//     if (SendSocket == INVALID_SOCKET)
//     {
//         wprintf(L"1 socket failed with error: %ld\n", WSAGetLastError());
//         WSACleanup();
//         return 1;
//     }
//     //---------------------------------------------
//     // Set up the RecvAddr structure with the IP address of
//     // the receiver (in this example case "192.168.1.1")
//     // and the specified port number.
//     RecvAddr.sin_family = AF_INET;
//     RecvAddr.sin_port = htons(Port);
//     RecvAddr.sin_addr.s_addr = inet_addr("192.168.1.1");
//     // ----------------------------------------------
//     // set IPPROTO_IP data not be fragmented
//     BOOL bOptVal = FALSE;
//     bOptVal = TRUE;
//     int bOptLen = sizeof(BOOL);
//     int R = 0;
//     R = setsockopt(SendSocket, IPPROTO_IP, IP_DONTFRAGMENT, (char *)&bOptVal, bOptLen);
//     if (R == SOCKET_ERROR)
//     {
//         wprintf(L"1error: %ld\n", WSAGetLastError());
//         return 1;
//     }
//     // set SOL_SOCKET SO_BROADCAST
//     BOOL cOptVal = TRUE;
//     INT cOptLen = sizeof(BOOL);
//     R = setsockopt(SendSocket, SOL_SOCKET, SO_BROADCAST, (char *)&cOptVal, cOptLen);
//     if (R == SOCKET_ERROR)
//     {
//         wprintf(L"2error: %ld\n", WSAGetLastError());
//         return 1;
//     }
//     //---------------------------------------------
//     // Send a datagram to the receiver
//     wprintf(L"Sending a datagram to the receiver...\n");
//     iResult = sendto(SendSocket,
//                      SendBuf, BufLen, 0, (SOCKADDR *)&RecvAddr, sizeof(RecvAddr));
//     if (iResult == SOCKET_ERROR)
//     {
//         wprintf(L"sendto failed with error: %d\n", WSAGetLastError());
//         closesocket(SendSocket);
//         WSACleanup();
//         return 1;
//     }
//     //---------------------------------------------
//     // When the application is finished sending, close the socket.
//     wprintf(L"Finished sending. Closing socket.\n");
//     iResult = closesocket(SendSocket);
//     if (iResult == SOCKET_ERROR)
//     {
//         wprintf(L"closesocket failed with error: %d\n", WSAGetLastError());
//         WSACleanup();
//         return 1;
//     }
//     //---------------------------------------------
//     // Clean up and quit.
//     wprintf(L"Exiting.\n");
//     WSACleanup();
//     return 0;
// }
