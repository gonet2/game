package main

import (
	"game/etcdclient"
	"game/numbers"
	pb "game/proto"
	"game/services"
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
	cli "gopkg.in/urfave/cli.v2"
)

func main() {
	app := &cli.App{
		Name: "agent",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "listen",
				Value: ":8888",
				Usage: "listening address:port",
			},
			&cli.StringSliceFlag{
				Name:  "etcd-hosts",
				Value: cli.NewStringSlice("http://127.0.0.1:2379"),
				Usage: "etcd hosts",
			},
			&cli.StringFlag{
				Name:  "etcd-root",
				Value: "/backends",
				Usage: "etcd root path",
			},
			&cli.StringFlag{
				Name:  "numbers",
				Value: "/numbers",
				Usage: "numbers path in etcd",
			},
			&cli.StringSliceFlag{
				Name:  "services",
				Value: cli.NewStringSlice("snowflake-10000"),
				Usage: "auto-discovering services",
			},
		},
		Action: func(c *cli.Context) error {
			log.Println("listen:", c.String("listen"))
			log.Println("etcd-hosts:", c.StringSlice("etcd-hosts"))
			log.Println("etcd-root:", c.String("etcd-root"))
			log.Println("services:", c.StringSlice("services"))
			log.Println("numbers:", c.String("numbers"))
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
			etcdclient.Init(c.StringSlice("etcd-hosts"))
			services.Init(c.String("etcd-root"), c.StringSlice("etcd-hosts"), c.StringSlice("services"))
			numbers.Init(c.String("numbers"))
			// 开始服务
			return s.Serve(lis)
		},
	}
	app.Run(os.Args)
}
