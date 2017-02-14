package etcdclient

import (
	log "github.com/Sirupsen/logrus"
	etcdclient "github.com/coreos/etcd/client"
)

var client etcdclient.Client

func Init(host []string) {
	// config
	cfg := etcdclient.Config{
		Endpoints: host,
		Transport: etcdclient.DefaultTransport,
	}

	// create client
	etcdcli, err := etcdclient.New(cfg)
	if err != nil {
		log.Panic(err)
		return
	}
	client = etcdcli
}

func KeysAPI() etcdclient.KeysAPI {
	return etcdclient.NewKeysAPI(client)
}

func NewOptions() etcdclient.GetOptions {
	return etcdclient.GetOptions{}
}

func NewWatcherOptions(recursive bool) *etcdclient.WatcherOptions {
	return &etcdclient.WatcherOptions{Recursive: recursive}
}
