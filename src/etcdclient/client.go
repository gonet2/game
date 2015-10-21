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

func init() {
	// etcd client
	machines = []string{DEFAULT_ETCD}
	if env := os.Getenv("ETCD_HOST"); env != "" {
		machines = strings.Split(env, ";")
	}
}

func KeysAPI() etcdclient.KeysAPI {
	cfg := etcdclient.Config{
		Endpoints: machines,
		Transport: etcdclient.DefaultTransport,
	}
	c, err := etcdclient.New(cfg)
	if err != nil {
		log.Critical(err)
		return nil
	}
	return etcdclient.NewKeysAPI(c)
}
