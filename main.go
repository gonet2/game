package main

import (
	"game/client_handler"
	"game/etcdclient"
	"game/kafka"
	"game/numbers"
	pb "game/proto"
	"game/services"
	"net"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
	cli "gopkg.in/urfave/cli.v2"
)

func main() {
	go http.ListenAndServe("0.0.0.0:6060", nil)
	app := &cli.App{
		Name:    "game",
		Usage:   "a stream processor based game",
		Version: "2.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "id",
				Value: "game1",
				Usage: "id of this service",
			},
			&cli.StringFlag{
				Name:  "listen",
				Value: ":10000",
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
				Name:  "kafka-brokers",
				Value: cli.NewStringSlice("127.0.0.1:9092"),
				Usage: "kafka brokers address",
			},
			&cli.StringFlag{
				Name:  "wal-topic",
				Value: "WAL-MYGAME",
				Usage: "write ahead log topic in kafka, very important!",
			},
			&cli.StringFlag{
				Name:  "trace-topic",
				Value: "trace-MYGAME",
				Usage: "user tracking topic",
			},
			&cli.StringSliceFlag{
				Name:  "services",
				Value: cli.NewStringSlice("snowflake-10000"),
				Usage: "auto-discovering services",
			},
			&cli.StringFlag{
				Name:  "mongodb",
				Value: "mongodb://127.0.0.1/mydb",
				Usage: "mongodb url",
			},
			&cli.DurationFlag{
				Name:  "mongodb-timeout",
				Value: 30 * time.Second,
				Usage: "mongodb socket timeout",
			},
			&cli.IntFlag{
				Name:  "mongodb-concurrent",
				Value: 128,
				Usage: "mongodb concurrent queries",
			},
		},
		Action: func(c *cli.Context) error {
			log.Println("id:", c.String("id"))
			log.Println("listen:", c.String("listen"))
			log.Println("etcd-hosts:", c.StringSlice("etcd-hosts"))
			log.Println("etcd-root:", c.String("etcd-root"))
			log.Println("services:", c.StringSlice("services"))
			log.Println("numbers:", c.String("numbers"))
			log.Println("kafka-brokers:", c.StringSlice("kafka-brokers"))
			log.Println("mongodb:", c.String("mongodb"))
			log.Println("mongodb-timeout:", c.Duration("mongodb-timeout"))
			log.Println("mongodb-concurrent:", c.Int("mongodb-concurrent"))

			// 监听
			lis, err := net.Listen("tcp", c.String("listen"))
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
			kafka.Init(c.StringSlice("kafka-brokers"), c.String("wal-topic"), c.String("trace-topic"), c.String("id"))
			client_handler.Init(c.String("mongodb"), c.Int("mongodb-concurrent"), c.Duration("mongodb-concurrent"))
			// 开始服务
			return s.Serve(lis)
		},
	}
	app.Run(os.Args)
}
