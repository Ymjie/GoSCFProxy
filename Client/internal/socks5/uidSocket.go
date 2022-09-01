package Socks5

import (
	"net"
	"sync"
)

type Conn struct {
	net.Conn
	Stats int // 0.未入FW  1. 入FW
}
type connmap map[string]Conn
type ConnList struct {
	sync.RWMutex
	Map connmap
}

var uidSocket = ConnList{sync.RWMutex{}, connmap{}}

func (l *ConnList) Num() int {
	return len(l.Map)
}

func (l *ConnList) GET(key string) (Conn, bool) {
	l.RLock()
	value, ok := l.Map[key]
	//value.Conn
	l.RUnlock()
	return value, ok
}

func (l *ConnList) ADD(key string, value net.Conn) {
	l.Lock()
	l.Map[key] = Conn{
		Conn:  value,
		Stats: 0,
	}
	l.Unlock()
}
func (l *ConnList) DEL(key string) {
	l.Lock()
	//l.Map[key] = value
	delete(l.Map, key)
	l.Unlock()
}
