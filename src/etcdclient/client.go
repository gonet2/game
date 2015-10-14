package etcdclient

import (
	"github.com/coreos/go-etcd/etcd"
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

func GetClient() *etcd.Client {
	return etcd.NewClient(machines)
}
