package main

import (
	"net"
	"os"
	pb "proto"

	log "github.com/gonet2/libs/nsq-logger"
	_ "github.com/gonet2/libs/statsd-pprof"
	"google.golang.org/grpc"
)

func main() {
	log.SetPrefix(SERVICE)
	// 监听
	lis, err := net.Listen("tcp", _port)
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
	log.Info("listening on ", lis.Addr())

	// 注册服务
	s := grpc.NewServer()
	ins := new(server)
	pb.RegisterGameServiceServer(s, ins)

	// 开始服务
	s.Serve(lis)
}
