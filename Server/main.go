package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

var (
	host_ = "0.0.0.0"
	port_ = 9000
)

func main() {
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		destAddrPort := r.PostForm.Get("destAddrPort")
		myaddr := r.PostForm.Get("myaddr")
		uid := r.PostForm.Get("uid")
		log.Printf("destAddrPort: %s BridgeAddrPort:%s uid:%s", destAddrPort, myaddr, uid)
		if len(destAddrPort) == 0 {
			log.Println("destAddrPort failed")
			rw.Write([]byte("NO!"))
			return
		}
		out, err := net.Dial("tcp", destAddrPort)
		if err != nil {
			return
		}
		bridge, err := net.Dial("tcp", myaddr)
		if err != nil {
			return
		}
		_, err = bridge.Write([]byte(uid))
		if err != nil {
			return
		}

		go Forward(out, bridge)
	})
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host_, port_), nil)
	if err != nil {
		log.Fatal("setup server fatal:", err)
	}
}

func Forward(client, target net.Conn) {
	forward := func(src, dest net.Conn) {
		defer src.Close()
		defer dest.Close()
		io.Copy(src, dest)
	}
	go forward(client, target)
	go forward(target, client)
}
