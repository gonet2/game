package services

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	etcdclient "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	DEFAULT_ETCD         = "http://172.17.42.1:2379"
	DEFAULT_SERVICE_PATH = "/backends"
	DEFAULT_NAME_FILE    = "/backends/names"
)

// a single connection
type client struct {
	key  string
	conn *grpc.ClientConn
}

// a kind of service
type service struct {
	clients []client
	idx     uint32 // for round-robin purpose
}

// all services
type service_pool struct {
	services          map[string]*service
	known_names       map[string]bool // store names.txt
	enable_name_check bool
	client            etcdclient.Client
	callbacks         map[string][]chan string // service add callback notify
	sync.RWMutex
}

var (
	_default_pool service_pool
	once          sync.Once
)

// Init() ***MUST*** be called before using
func Init(names ...string) {
	once.Do(func() { _default_pool.init(names...) })
}

func (p *service_pool) init(names ...string) {
	// etcd client
	machines := []string{DEFAULT_ETCD}
	if env := os.Getenv("ETCD_HOST"); env != "" {
		machines = strings.Split(env, ";")
	}

	// init etcd client
	cfg := etcdclient.Config{
		Endpoints: machines,
		Transport: etcdclient.DefaultTransport,
	}
	c, err := etcdclient.New(cfg)
	if err != nil {
		log.Panic(err)
		os.Exit(-1)
	}
	p.client = c

	// init
	p.services = make(map[string]*service)
	p.known_names = make(map[string]bool)

	// names init
	if len(names) == 0 { // names not provided
		names = p.load_names() // try read from names.txt
	}
	if len(names) > 0 {
		p.enable_name_check = true
	}

	log.Println("all service names:", names)
	for _, v := range names {
		p.known_names[DEFAULT_SERVICE_PATH+"/"+strings.TrimSpace(v)] = true
	}

	// start connection
	p.connect_all(DEFAULT_SERVICE_PATH)
}

// get stored service name
func (p *service_pool) load_names() []string {
	kAPI := etcdclient.NewKeysAPI(p.client)
	// get the keys under directory
	log.Println("reading names:", DEFAULT_NAME_FILE)
	resp, err := kAPI.Get(context.Background(), DEFAULT_NAME_FILE, nil)
	if err != nil {
		log.Println(err)
		return nil
	}

	// validation check
	if resp.Node.Dir {
		log.Println("names is not a file")
		return nil
	}

	// split names
	return strings.Split(resp.Node.Value, "\n")
}

// connect to all services
func (p *service_pool) connect_all(directory string) {
	kAPI := etcdclient.NewKeysAPI(p.client)
	// get the keys under directory
	log.Println("connecting services under:", directory)
	resp, err := kAPI.Get(context.Background(), directory, &etcdclient.GetOptions{Recursive: true})
	if err != nil {
		log.Println(err)
		return
	}

	// validation check
	if !resp.Node.Dir {
		log.Println("not a directory")
		return
	}

	for _, node := range resp.Node.Nodes {
		if node.Dir { // service directory
			for _, service := range node.Nodes {
				p.add_service(service.Key, service.Value)
			}
		}
	}
	log.Println("services add complete")

	go p.watcher()
}

// watcher for data change in etcd directory
func (p *service_pool) watcher() {
	kAPI := etcdclient.NewKeysAPI(p.client)
	w := kAPI.Watcher(DEFAULT_SERVICE_PATH, &etcdclient.WatcherOptions{Recursive: true})
	for {
		resp, err := w.Next(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}
		if resp.Node.Dir {
			continue
		}

		switch resp.Action {
		case "set", "create", "update", "compareAndSwap":
			p.add_service(resp.Node.Key, resp.Node.Value)
		case "delete":
			p.remove_service(resp.PrevNode.Key)
		}
	}
}

// add a service
func (p *service_pool) add_service(key, value string) {
	p.Lock()
	defer p.Unlock()
	// name check
	service_name := filepath.Dir(key)
	if p.enable_name_check && !p.known_names[service_name] {
		return
	}

	// try new service kind init
	if p.services[service_name] == nil {
		p.services[service_name] = &service{}
	}

	// create service connection
	service := p.services[service_name]
	if conn, err := grpc.Dial(value, grpc.WithBlock(), grpc.WithInsecure()); err == nil {
		service.clients = append(service.clients, client{key, conn})
		log.Println("service added:", key, "-->", value)
		for k := range p.callbacks[service_name] {
			select {
			case p.callbacks[service_name][k] <- key:
			default:
			}
		}
	} else {
		log.Println("did not connect:", key, "-->", value, "error:", err)
	}
}

// remove a service
func (p *service_pool) remove_service(key string) {
	p.Lock()
	defer p.Unlock()
	// name check
	service_name := filepath.Dir(key)
	if p.enable_name_check && !p.known_names[service_name] {
		return
	}

	// check service kind
	service := p.services[service_name]
	if service == nil {
		log.Println("no such service:", service_name)
		return
	}

	// remove a service
	for k := range service.clients {
		if service.clients[k].key == key { // deletion
			service.clients[k].conn.Close()
			service.clients = append(service.clients[:k], service.clients[k+1:]...)
			log.Println("service removed:", key)
			return
		}
	}
}

// provide a specific key for a service, eg:
// path:/backends/snowflake, id:s1
//
// the full cannonical path for this service is:
// 			/backends/snowflake/s1
func (p *service_pool) get_service_with_id(path string, id string) *grpc.ClientConn {
	p.RLock()
	defer p.RUnlock()
	// check existence
	service := p.services[path]
	if service == nil {
		return nil
	}
	if len(service.clients) == 0 {
		return nil
	}

	// loop find a service with id
	fullpath := string(path) + "/" + id
	for k := range service.clients {
		if service.clients[k].key == fullpath {
			return service.clients[k].conn
		}
	}

	return nil
}

// get a service in round-robin style
// especially useful for load-balance with state-less services
func (p *service_pool) get_service(path string) (conn *grpc.ClientConn, key string) {
	p.RLock()
	defer p.RUnlock()
	// check existence
	service := p.services[path]
	if service == nil {
		return nil, ""
	}

	if len(service.clients) == 0 {
		return nil, ""
	}

	// get a service in round-robind style,
	idx := int(atomic.AddUint32(&service.idx, 1)) % len(service.clients)
	return service.clients[idx].conn, service.clients[idx].key
}

func (p *service_pool) register_callback(path string, callback chan string) {
	p.Lock()
	defer p.Unlock()
	if p.callbacks == nil {
		p.callbacks = make(map[string][]chan string)
	}

	p.callbacks[path] = append(p.callbacks[path], callback)
	if s, ok := p.services[path]; ok {
		for k := range s.clients {
			callback <- s.clients[k].key
		}
	}
	log.Println("register callback on:", path)
}

/////////////////////////////////////////////////////////////////
// Wrappers
func GetService(path string) *grpc.ClientConn {
	conn, _ := _default_pool.get_service(path)
	return conn
}

func GetService2(path string) (*grpc.ClientConn, string) {
	conn, key := _default_pool.get_service(path)
	return conn, key
}

func GetServiceWithId(path string, id string) *grpc.ClientConn {
	return _default_pool.get_service_with_id(path, id)
}

func RegisterCallback(path string, callback chan string) {
	_default_pool.register_callback(path, callback)
}
