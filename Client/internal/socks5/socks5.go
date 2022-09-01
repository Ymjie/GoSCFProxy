package Socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Ymjie/GoSCFProxy/Client/internal/config"
	Log "github.com/Ymjie/GoSCFProxy/Client/pkg/logger"
	"github.com/google/uuid"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

func Start(C *config.Config) {
	ch := make(chan int)
	logger := C.Log
	addr := fmt.Sprintf("%v:%v", C.Listener.Socks5.ListenIP, C.Listener.Socks5.ListenPort)
	logger.PF(Log.LINFO, "<Socks5>Listen: %v\n", addr)
	server, err := net.Listen("tcp", addr)
	if err != nil {
		logger.PF(Log.LERROR, "Listen failed: %v\n", err)
		return
	}

	go SCF_connect(ch, C)
	go Accepthandle(server, C)
	select {
	case <-ch:
		logger.PF(Log.LINFO, "Exit!")
	}
}

func Accepthandle(server net.Listener, C *config.Config) {
	logger := C.Log
	for {
		client, err := server.Accept()
		logger.PF(Log.LDEBUG, "<Socks5>a connecting\n")
		if err != nil {
			logger.PF(Log.LERROR, "<Socks5>ERR:%v\n", err)
			continue
		}
		go process(C, client)
	}
}
func Socks5Auth(client net.Conn) (err error) {
	buf := make([]byte, 256)

	// 读取 VER 和 NMETHODS
	n, err := io.ReadFull(client, buf[:2])
	if n != 2 {
		return errors.New("reading header: " + err.Error())
	}

	ver, nMethods := int(buf[0]), int(buf[1])
	if ver != 5 {
		return errors.New("invalid version")
	}

	// 读取 METHODS 列表
	n, err = io.ReadFull(client, buf[:nMethods])
	if n != nMethods {
		return errors.New("reading methods: " + err.Error())
	}

	//无需认证
	n, err = client.Write([]byte{0x05, 0x00})
	if n != 2 || err != nil {
		return errors.New("write rsp: " + err.Error())
	}

	return nil
}

func Socks5Connect(client net.Conn) (string, error) {
	buf := make([]byte, 256)

	n, err := io.ReadFull(client, buf[:4])
	if n != 4 {
		return "", errors.New("read header: " + err.Error())
	}

	ver, cmd, _, atyp := buf[0], buf[1], buf[2], buf[3]
	if ver != 5 || cmd != 1 {
		return "", errors.New("invalid ver/cmd")
	}

	addr := ""
	switch atyp {
	case 1:
		n, err = io.ReadFull(client, buf[:4])
		if n != 4 {
			return "", errors.New("invalid IPv4: " + err.Error())
		}
		addr = fmt.Sprintf("%d.%d.%d.%d", buf[0], buf[1], buf[2], buf[3])

	case 3:
		n, err = io.ReadFull(client, buf[:1])
		if n != 1 {
			return "", errors.New("invalid hostname: " + err.Error())
		}
		addrLen := int(buf[0])

		n, err = io.ReadFull(client, buf[:addrLen])
		if n != addrLen {
			return "", errors.New("invalid hostname: " + err.Error())
		}
		addr = string(buf[:addrLen])

	case 4:
		return "", errors.New("IPv6: no supported yet")

	default:
		return "", errors.New("invalid atyp")
	}

	n, err = io.ReadFull(client, buf[:2])
	if n != 2 {
		return "", errors.New("read port: " + err.Error())
	}
	port := binary.BigEndian.Uint16(buf[:2])

	destAddrPort := fmt.Sprintf("%s:%d", addr, port)

	//dest, err := net.Dial("tcp", destAddrPort)
	//if err != nil {
	//	return nil, errors.New("dial dst: " + err.Error())
	//}

	n, err = client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err != nil {
		//dest.Close()
		return "", errors.New("write rsp: " + err.Error())
	}

	return destAddrPort, nil
}
func process(C *config.Config, client net.Conn) {
	logger := C.Log
	poll := C.Poll
	if err := Socks5Auth(client); err != nil {
		logger.PF(Log.LWARN, "<Socks5> auth error:%v\n", err)
		client.Close()
		return
	}
	destAddrPort, err := Socks5Connect(client)
	if err != nil {
		logger.PF(Log.LWARN, "<Socks5> connect error:%v\n", err)
		client.Close()
		return
	}
	uid := uuid.New().String()

	uidSocket.ADD(uid, client)

	if len(C.Listener.Bridge.IP) == 0 {
		ip, err := Getip()
		if err != nil {
			C.Log.PF(Log.LERROR, "<Bridge>GET IP Addr err:%v\n", err)
		}
		C.Listener.Bridge.IP = ip
	}
	addr := net.ParseIP(C.Listener.Bridge.IP)
	if addr == nil {
		C.Log.PF(Log.LFATAL, "Listener.Bridge.IP error:%v\n", C.Listener.Bridge.IP)
		return
	}
	myaddr := fmt.Sprintf("%v:%v", addr, C.Listener.Bridge.Port)

	data := url.Values{"destAddrPort": {destAddrPort}, "uid": {uid}, "myaddr": {myaddr}}
	C.Log.PF(Log.LDEBUG, "POST:%v\n", data)
	SCFurl, err := poll.Get(destAddrPort)
	if err != nil {
		logger.PF(Log.LFATAL, "<Socks5>Get SCFurl err:%v\n", err)
		uidSocket.DEL(uid)
		return
	}
	if len(SCFurl) == 0 {
		logger.PF(Log.LFATAL, "<Socks5>no SCFurl\n")
		uidSocket.DEL(uid)
		return
	}
	_, err = http.PostForm(SCFurl, data)
	if err != nil {
		logger.PF(Log.LFATAL, "<Socks5>post fatal:%v\n", err)
		uidSocket.DEL(uid)
	}
	go func(uid string) {
		<-time.After(15 * time.Second)
		if conn, ok := uidSocket.GET(uid); conn.Stats == 0 && ok {
			uidSocket.DEL(uid)
			logger.PF(Log.LDEBUG, "<Socks5>uid:%v,超时15s自动释放.\n", uid)
		}
	}(uid)

}
