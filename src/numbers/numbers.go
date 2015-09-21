package numbers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	log "github.com/gonet2/nsq-logger"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DEFAULT_NUMBERS_PATH = "/numbers"
	DEFAULT_ETCD         = "http://172.17.42.1:2379"
	RETRY_DELAY          = 10 * time.Second
)

var (
	_default_numbers numbers
)

func init() {
	_default_numbers.init()
	_default_numbers.load(DEFAULT_NUMBERS_PATH)
	go _default_numbers.watcher()
}

// a record contains fields
type record struct {
	fields map[string]string
}

// a table contains records
type table struct {
	records map[string]*record // all records
	keys    []string           // all record keys
}

// numbers contains all tables
type numbers struct {
	tables      map[string]*table
	client_pool sync.Pool // etcd client pool
	sync.RWMutex
}

func (ns *numbers) init() {
	machines := []string{DEFAULT_ETCD}
	if env := os.Getenv("ETCD_HOST"); env != "" {
		machines = strings.Split(env, ";")
	}

	ns.client_pool.New = func() interface{} {
		return etcd.NewClient(machines)
	}
}

// load all csv(s) from a given directory in etcd
func (ns *numbers) load(directory string) {
	client := ns.client_pool.Get().(*etcd.Client)
	defer func() {
		ns.client_pool.Put(client)
	}()

	// get the keys under directory
	log.Info("loading numbers from:", directory)
	resp, err := client.Get(directory, true, false)
	if err != nil {
		log.Error(err)
		return
	}

	// validation check
	if !resp.Node.Dir {
		log.Error("not a directory")
		return
	}

	// read all csv(s) from this dirctory
	ns.tables = make(map[string]*table)
	count := 0
	for _, node := range resp.Node.Nodes {
		if !node.Dir {
			log.Trace("loading:", node.Key)
			ns.parse(node.Key, node.Value)
			count++
		} else {
			log.Warning("not a file:", node.Key)
		}
	}

	log.Finef("%v csv(s) loaded", count)
}

// watcher for data change in etcd directory
func (ns *numbers) watcher() {
	client := ns.client_pool.Get().(*etcd.Client)
	defer func() {
		ns.client_pool.Put(client)
	}()

	for {
		ch := make(chan *etcd.Response, 10)
		go func() {
			for {
				if resp, ok := <-ch; ok {
					if !resp.Node.Dir {
						ns.parse(resp.Node.Key, resp.Node.Value)
						log.Trace("csv change:", resp.Node.Key)
					}
				} else {
					return
				}
			}
		}()

		_, err := client.Watch(DEFAULT_NUMBERS_PATH, 0, true, ch, nil)
		if err != nil {
			log.Critical(err)
		}
		<-time.After(RETRY_DELAY)
	}
}

// parse & load a csv
func (ns *numbers) parse(key, value string) {
	ns.Lock()
	defer ns.Unlock()
	src := bytes.NewBufferString(value)
	csv_reader := csv.NewReader(src)
	records, err := csv_reader.ReadAll()
	if err != nil {
		log.Errorf("%v %v", err, key)
		return
	}

	if len(records) == 0 {
		log.Warningf("empty document: %v", key)
		return
	}

	tblname := filepath.Base(key)
	// 记录数据, 第一行为表头，因此从第二行开始
	for line := 1; line < len(records); line++ {
		for field := 1; field < len(records[line]); field++ { // 每条记录的第一个字段作为行索引
			ns.set(ns.tables, tblname, records[line][0], records[0][field], records[line][field])
		}
	}

	// 记录KEYS
	ns.dump_keys(ns.tables, tblname)
}

// set field value
func (ns *numbers) set(tables map[string]*table, tblname string, rowname string, fieldname string, value string) {
	tbl, ok := tables[tblname]
	if !ok {
		tbl = &table{}
		tbl.records = make(map[string]*record)
		tables[tblname] = tbl
	}

	rec, ok := tbl.records[rowname]
	if !ok {
		rec = &record{}
		rec.fields = make(map[string]string)
		tbl.records[rowname] = rec
	}

	rec.fields[fieldname] = value
}

// dump keys
func (ns *numbers) dump_keys(tables map[string]*table, tblname string) {
	tbl, ok := tables[tblname]
	if !ok {
		panic(fmt.Sprint("table ", tblname, " not exists!"))
	}

	for k := range tbl.records {
		tbl.keys = append(tbl.keys, k)
	}
}

// get field value
func (ns *numbers) get(tblname string, rowname string, fieldname string) string {
	ns.RLock()
	defer ns.RUnlock()
	tables := ns.tables

	tbl, ok := tables[tblname]
	if !ok {
		panic(fmt.Sprint("table ", tblname, " not exists!"))
	}

	rec, ok := tbl.records[rowname]
	if !ok {
		panic(fmt.Sprint("table ", tblname, " row ", rowname, " not exists!"))
	}

	value, ok := rec.fields[fieldname]
	if !ok {
		panic(fmt.Sprint("table ", tblname, " field ", fieldname, " not exists!"))
	}
	return value
}

// get field value as integer
func (ns *numbers) GetInt(tblname string, rowname interface{}, fieldname string) int32 {
	val := ns.get(tblname, fmt.Sprint(rowname), fieldname)
	if val == "" {
		return 0
	}

	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		panic(fmt.Sprintf("cannot parse integer from gamedata %v %v %v %v\n", tblname, rowname, fieldname, err))
	}

	// round to the integer
	// 1.00000001 -> 1
	// 0.99999999 -> 1
	// -0.9999990 -> -1
	// -1.0000001 -> -1
	return int32(v)
}

// get field value as float
func (ns *numbers) GetFloat(tblname string, rowname interface{}, fieldname string) float64 {
	val := ns.get(tblname, fmt.Sprint(rowname), fieldname)
	if val == "" {
		return 0.0
	}

	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		panic(fmt.Sprintf("cannot parse float from gamedata %v %v %v %v\n", tblname, rowname, fieldname, err))
	}

	return f
}

// get field value as string
func (ns *numbers) GetString(tblname string, rowname interface{}, fieldname string) string {
	return ns.get(tblname, fmt.Sprint(rowname), fieldname)
}

// get all keys
func (ns *numbers) GetKeys(tblname string) []string {
	ns.RLock()
	defer ns.RUnlock()
	tables := ns.tables

	tbl, ok := tables[tblname]
	if !ok {
		panic(fmt.Sprint("table ", tblname, " not exists!"))
	}

	return tbl.keys
}

// get row count
func (ns *numbers) Count(tblname string) int32 {
	ns.RLock()
	defer ns.RUnlock()
	tables := ns.tables

	tbl, ok := tables[tblname]
	if !ok {
		panic(fmt.Sprint("table ", tblname, " not exists!"))
	}

	return int32(len(tbl.records))
}

// test record exists
func (ns *numbers) IsRecordExists(tblname string, rowname interface{}) bool {
	ns.RLock()
	defer ns.RUnlock()
	tables := ns.tables

	tbl, ok := tables[tblname]
	if !ok {
		return false
	}

	_, ok = tbl.records[fmt.Sprint(rowname)]
	if !ok {
		return false
	}

	return true
}

// test field exists
func (ns *numbers) IsFieldExists(tblname string, fieldname string) bool {
	ns.RLock()
	defer ns.RUnlock()
	tables := ns.tables

	// check table existence
	tbl, ok := tables[tblname]
	if !ok {
		return false
	}

	// get one record key
	key := ""
	for k, _ := range tbl.records {
		key = k
		break
	}

	rec, ok := tbl.records[key]
	if !ok {
		return false
	}

	_, ok = rec.fields[fieldname]
	if !ok {
		return false
	}

	return true
}

func GetInt(tblname string, rowname interface{}, fieldname string) int32 {
	return _default_numbers.GetInt(tblname, rowname, fieldname)
}

func GetFloat(tblname string, rowname interface{}, fieldname string) float64 {
	return _default_numbers.GetFloat(tblname, rowname, fieldname)
}

func GetString(tblname string, rowname interface{}, fieldname string) string {
	return _default_numbers.GetString(tblname, rowname, fieldname)
}

func GetKeys(tblname string) []string {
	return _default_numbers.GetKeys(tblname)
}

func Count(tblname string) int32 {
	return _default_numbers.Count(tblname)
}

func IsFieldExists(tblname string, fieldname string) bool {
	return _default_numbers.IsFieldExists(tblname, fieldname)
}

func IsRecordExists(tblname string, rowname interface{}) bool {
	return _default_numbers.IsRecordExists(tblname, rowname)
}
