package numbers

import (
	"encoding/base64"
	"fmt"
	gopkg "path"
	"strconv"
	"sync"

	//	cli "gopkg.in/urfave/cli.v2"

	"game/etcdclient"

	log "github.com/Sirupsen/logrus"
	"github.com/tealeg/xlsx"
	"golang.org/x/net/context"
)

type NumbersOp interface {
	GetInt(tblname string, rowname interface{}, fieldname string) int32
	GetFloat(tblname string, rowname interface{}, fieldname string) float64
	GetString(tblname string, rowname interface{}, fieldname string) string
	GetKeys(tblname string) []string
	IsFieldExists(tblname string, rowname interface{}, fieldname string) bool
	IsRecordExists(tblname string, rowname interface{}) bool
	IsTableExists(tblname string) bool
}

var (
	_dataConfig configs
)

func Init(path string) {
	_dataConfig.init(path)
	go watcher(path)
}

// 字段定义
type record struct {
	fields map[string]string
}

// 表定义
type table struct {
	records map[string]*record
	keys    []string
}

// numbers可以读取以下的结构的xlsx sheet
// (null)   字段1    字段2    字段3
// 记录1    值       值       值
// 记录2    值       值       值
// 记录3    值       值       值

// read only
// 数值类
type numbers struct {
	tables map[string]*table
	name   string
}

type configs struct {
	numbers map[string]*numbers
	sync.RWMutex
}

// Numbers 获取excel
func Numbers(name string) NumbersOp {
	_dataConfig.RLock()
	defer _dataConfig.RUnlock()
	if n, ok := _dataConfig.numbers[name]; ok {
		return NumbersOp(n)
	}
	panic(fmt.Sprintf("numbers not exists %v", name))
}

// SetNumber 更新表
func SetNumbers(ns *numbers) {
	_dataConfig.Lock()
	defer _dataConfig.Unlock()
	if _dataConfig.numbers == nil {
		_dataConfig.numbers = make(map[string]*numbers)
	}
	_dataConfig.numbers[ns.name] = ns
}

func (c *configs) init(path string) {
	kapi := etcdclient.KeysAPI()
	opt := etcdclient.NewOptions()
	resp, err := kapi.Get(context.Background(), path, &opt)
	if err != nil {
		log.Error(err)
		return
	}

	for i := range resp.Node.Nodes {
		node := resp.Node.Nodes[i]
		// 解码xlsx
		xlsx_bin, err := base64.StdEncoding.DecodeString(node.Value)
		if err != nil {
			log.Error(err, node.Key)
			return
		}

		// 读取xlsx
		xlsx_reader, err := xlsx.OpenBinary(xlsx_bin)
		if err != nil {
			log.Error(err, node.Key)
			return
		}
		ns := &numbers{tables: make(map[string]*table), name: gopkg.Base(node.Key)}
		ns.parse(gopkg.Base(node.Key), xlsx_reader.Sheets)
		SetNumbers(ns)
	}
}

// 载入数据
func (ns *numbers) parse(xlsxname string, sheets []*xlsx.Sheet) {
	var sheetName string
	defer func() {
		if x := recover(); x != nil {
			log.WithField("errmsg", fmt.Sprintf("xls %v sheetName %v err %v", xlsxname, sheetName, x)).
				WithField("err", x).
				Error()
		}
	}()

	for _, sheet := range sheets {
		//	println("parse sheet ", sheet.Name)
		// 第一行为表头，因此从第二行开始
		if len(sheet.Rows) > 0 {
			header := sheet.Rows[0]
			for i := 1; i < len(sheet.Rows); i++ {
				row := sheet.Rows[i]
				for j := 0; j < len(row.Cells); j++ {
					rowname, _ := row.Cells[0].String()
					fieldname, _ := header.Cells[j].String()
					value, _ := row.Cells[j].String()
					ns.set(sheet.Name, rowname, fieldname, value)
				}
			}
		}
		ns.dump_keys(sheet.Name)
	}
}

// 设置值
func (ns *numbers) set(tblname string, rowname string, fieldname string, value string) {
	tbl, ok := ns.tables[tblname]
	if !ok {
		tbl = &table{}
		tbl.records = make(map[string]*record)
		ns.tables[tblname] = tbl
	}

	rec, ok := tbl.records[rowname]
	if !ok {
		rec = &record{}
		rec.fields = make(map[string]string)
		tbl.records[rowname] = rec
	}

	rec.fields[fieldname] = value
}

// 记录所有的KEY
func (ns *numbers) dump_keys(tblname string) {
	tbl, ok := ns.tables[tblname]
	if !ok {
		panic(fmt.Sprint("table ", tblname, " not exists!"))
	}

	for k := range tbl.records {
		tbl.keys = append(tbl.keys, k)
	}
}

// 读取值
func (ns *numbers) get(tblname string, rowname string, fieldname string) string {
	tbl, ok := ns.tables[tblname]
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

func (ns *numbers) IsTableExists(tblname string) bool {
	_, ok := ns.tables[tblname]
	return ok
}

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

func (ns *numbers) GetString(tblname string, rowname interface{}, fieldname string) string {
	return ns.get(tblname, fmt.Sprint(rowname), fieldname)
}

func (ns *numbers) GetKeys(tblname string) []string {
	tbl, ok := ns.tables[tblname]
	if !ok {
		panic(fmt.Sprint("table ", tblname, " not exists!"))
	}

	return tbl.keys
}

func (ns *numbers) Count(tblname string) int32 {
	tbl, ok := ns.tables[tblname]
	if !ok {
		panic(fmt.Sprint("table ", tblname, " not exists!"))
	}

	return int32(len(tbl.records))
}

func (ns *numbers) IsRecordExists(tblname string, rowname interface{}) bool {
	tbl, ok := ns.tables[tblname]
	if !ok {
		return false
	}

	_, ok = tbl.records[fmt.Sprint(rowname)]
	if !ok {
		return false
	}

	return true
}

func (ns *numbers) IsFieldExists(tblname string, rowname interface{}, fieldname string) bool {
	tbl, ok := ns.tables[tblname]
	if !ok {
		return false
	}

	key := fmt.Sprint(rowname)
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

func watcher(path string) {
	kapi := etcdclient.KeysAPI()
	w := kapi.Watcher(path, etcdclient.NewWatcherOptions(true))
	for {
		resp, err := w.Next(context.Background())
		if err != nil {
			log.Error(err)
			continue
		}
		log.Info(resp)
		switch resp.Action {
		case "set", "create", "update", "compareAndSwap":
			xlsx_bin, err := base64.StdEncoding.DecodeString(resp.Node.Value)
			if err != nil {
				log.Error(err)
				continue
			}

			// 读取xlsx
			xlsx_reader, err := xlsx.OpenBinary(xlsx_bin)
			if err != nil {
				log.Error(err)
				continue
			}
			ns := &numbers{tables: make(map[string]*table), name: gopkg.Base(resp.Node.Key)}
			ns.parse(gopkg.Base(resp.Node.Key), xlsx_reader.Sheets)
			SetNumbers(ns)
		}
	}
}
