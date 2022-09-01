package Socks5

import (
	"fmt"
	"github.com/Ymjie/GoSCFProxy/Client/internal/config"
	Log "github.com/Ymjie/GoSCFProxy/Client/pkg/logger"
	"io"
	"net"
)

func SCF_connect(done chan int, C *config.Config) {
	if len(C.Listener.Bridge.ListenIP) == 0 {
		C.Listener.Bridge.ListenIP = "0.0.0.0"
	}
	addr := net.ParseIP(C.Listener.Bridge.ListenIP)

	if addr == nil {
		C.Log.PF(Log.LFATAL, "Bridge ip error:%v\n", C.Listener.Bridge.ListenIP)
		close(done)
		return
	}
	BridgeAddrPort := fmt.Sprintf("%v:%v", addr, C.Listener.Bridge.ListenPort)
	C.Log.PF(Log.LINFO, "<Bridge>Listen%v\n", BridgeAddrPort)
	server, err := net.Listen("tcp", BridgeAddrPort)
	if err != nil {
		C.Log.PF(Log.LWARN, "<Bridge>Listen failed: %v\n", err)
		close(done)
		return
	}

	for {
		client, err := server.Accept()
		C.Log.PF(Log.LDEBUG, "<Bridge>a connecting\n")
		if err != nil {
			C.Log.PF(Log.LWARN, "<Bridge>Accept failed: %v\n", err)
			continue
		}
		go func(bridge net.Conn) {
			C.Log.PF(Log.LINFO, "<Bridge>Get start data\n")
			buf := make([]byte, 256)
			n, err := io.ReadFull(bridge, buf[:36])
			if n != 36 {
				C.Log.PF(Log.LWARN, "<Bridge>GET UUID failed: %v\n", err)
				return
			}
			C.Log.PF(Log.LINFO, "<Bridge>uid obtained successfully\n")
			uid := string(buf[0:36])
			client, ok := uidSocket.GET(uid)

			if !ok {
				C.Log.PF(Log.LERROR, "<Bridge>GET conn with UUID failed")
				return
			}
			bridge_addr := bridge.RemoteAddr().String()
			client_addr := client.RemoteAddr().String()

			C.Log.PF(Log.LINFO, "<Bridge>%v:通道建立：%s<==>bridge_addr: %s 当前数量：%v \n", uid, client_addr, bridge_addr, uidSocket.Num())
			client.Stats = 1
			Forward(uid, client, bridge)

		}(client)
	}
}

func Forward(uid string, client, target net.Conn) {
	forward := func(src, dest net.Conn) {
		defer func() {
			src.Close()
			dest.Close()
			uidSocket.DEL(uid)
		}()
		io.Copy(src, dest)
	}
	go forward(client, target)
	go forward(target, client)
}
