package main

import (
	"github.com/liuyp5181/base/cache"
	"github.com/liuyp5181/base/config"
	"github.com/liuyp5181/base/database"
	"github.com/liuyp5181/base/service"
	pb "github.com/liuyp5181/gateway/api"
	"github.com/liuyp5181/gateway/handler"
)

func init() {
	config.ServiceName = pb.Greeter_ServiceDesc.ServiceName
}

func main() {
	err := cache.Connect("test")
	if err != nil {
		panic(err)
	}

	err = database.Connect("test")
	if err != nil {
		panic(err)
	}

	err = service.InitClients()
	if err != nil {
		panic(err)
	}

	panic(handler.Run(config.GetConfig().Server.IP, config.GetConfig().Server.Port))
}
