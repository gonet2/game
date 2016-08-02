package main

import (
	pb "game/proto"
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	sp "github.com/gonet2/libs/services"
	_ "github.com/gonet2/libs/statsd-pprof"
	"google.golang.org/grpc"
)

func main() {
	// 监听
	lis, err := net.Listen("tcp", _port)
	if err != nil {
		log.Panic(err)
		os.Exit(-1)
	}
	log.Info("listening on ", lis.Addr())

	// 注册服务
	s := grpc.NewServer()
	ins := new(server)
	pb.RegisterGameServiceServer(s, ins)

	// 初始化Services
	sp.Init("snowflake")
	// 开始服务
	s.Serve(lis)
}
