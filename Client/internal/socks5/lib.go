package Socks5

import (
	"io"
	"net/http"
)

func Getip() (string, error) {
	resp, err := http.Get("http://ipinfo.io/ip")
	if err != nil {
		return "", err
	}
	ip, _ := io.ReadAll(resp.Body)
	return string(ip), nil

}
