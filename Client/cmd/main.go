package main

import (
	"github.com/Ymjie/GoSCFProxy/Client/internal/config"
	Socks5 "github.com/Ymjie/GoSCFProxy/Client/internal/socks5"
	"github.com/Ymjie/GoSCFProxy/Client/pkg/load_balance"
	"github.com/Ymjie/GoSCFProxy/Client/pkg/logger"
	"log"
)

func main() {

	appconfig, err := config.Load("./config.yml")
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	poll := load_balance.LoadBalanceFactory(load_balance.LbRandom)
	err = poll.Add(appconfig.SCFList...)
	if err != nil {
		panic(err)
	}
	Mlog := logger.New(nil, appconfig.Logger.Level, 0)
	appconfig.Poll = poll
	appconfig.Log = Mlog
	Socks5.Start(appconfig)
}
