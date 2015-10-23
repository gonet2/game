package etcdclient

import (
	etcdclient "github.com/coreos/etcd/client"
	log "github.com/gonet2/libs/nsq-logger"
	"os"
	"strings"
)

const (
	DEFAULT_ETCD = "http://172.17.42.1:2379"
)

var machines []string
var client etcdclient.Client

func init() {
	// etcd client
	machines = []string{DEFAULT_ETCD}
	if env := os.Getenv("ETCD_HOST"); env != "" {
		machines = strings.Split(env, ";")
	}

	// config
	cfg := etcdclient.Config{
		Endpoints: machines,
		Transport: etcdclient.DefaultTransport,
	}

	// create client
	c, err := etcdclient.New(cfg)
	if err != nil {
		log.Critical(err)
		return
	}
	client = c
}

func KeysAPI() etcdclient.KeysAPI {
	return etcdclient.NewKeysAPI(client)
}
