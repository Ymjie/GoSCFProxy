package config

import (
	"github.com/Ymjie/GoSCFProxy/Client/pkg/load_balance"
	"github.com/Ymjie/GoSCFProxy/Client/pkg/logger"
)

type Config struct {
	Logger   Logger   `yaml:"logger"`
	Listener Listener `yaml:"Listener"`
	SCFList  []string `yaml:"SCFList"`
	Log      *logger.MyLogger
	Poll     load_balance.LoadBalance
}

type Logger struct {
	Level int64  `yaml:"level"`
	File  string `yaml:"file"`
}

type Listener struct {
	Socks5 Socks5 `yaml:"socks5"`
	Bridge Bridge `yaml:"bridge"`
}

type Socks5 struct {
	ListenPort int    `yaml:"ListenPort"`
	User       string `yaml:"user"`
	Passwd     string `yaml:"passwd"`
	Version    int    `yaml:"version"`
	ListenIP   string `yaml:"ListenIP"`
}

type Bridge struct {
	IP         string `yaml:"IP"`
	Port       string `yaml:"Port"`
	ListenIP   string `yaml:"ListenIP"`
	ListenPort int    `yaml:"ListenPort"`
}
