package numbers

import (
	"encoding/base64"
	"fmt"
	log "github.com/gonet2/libs/nsq-logger"
	"github.com/tealeg/xlsx"
	"golang.org/x/net/context"
	"strconv"
)

import (
	"etcdclient"
)

const (
	DEFAULT_NUMBERS_PATH = "/numbers"
	DEFAULT_ETCD         = "http://172.17.42.1:2379"
)

var (
	_default_numbers numbers
)

func init() {
	_default_numbers.init(DEFAULT_NUMBERS_PATH)
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

// 数值类
type numbers struct {
	tables map[string]*table
}

func (ns *numbers) init(path string) {
	ns.tables = make(map[string]*table)
	kapi := etcdclient.KeysAPI()
	resp, err := kapi.Get(context.Background(), path, nil)
	if err != nil {
		log.Error(err)
		return
	}

	// 解码xlsx
	xlsx_bin, err := base64.StdEncoding.DecodeString(resp.Node.Value)
	if err != nil {
		log.Critical(err)
		return
	}

	// 读取xlsx
	xlsx_reader, err := xlsx.OpenBinary(xlsx_bin)
	if err != nil {
		log.Critical(err)
		return
	}
	ns.parse(xlsx_reader.Sheets)
}

// 载入数据
func (ns *numbers) parse(sheets []*xlsx.Sheet) {
	for _, sheet := range sheets {
		// 第一行为表头，因此从第二行开始
		if len(sheet.Rows) > 0 {
			header := sheet.Rows[0]
			for i := 1; i < len(sheet.Rows); i++ {
				row := sheet.Rows[i]
				for j := 1; j < len(row.Cells); j++ {
					ns.set(sheet.Name, row.Cells[0].String(), header.Cells[j].String(), row.Cells[j].String())
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

func (ns *numbers) IsFieldExists(tblname string, fieldname string) bool {
	tbl, ok := ns.tables[tblname]
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

// 读取表tblname, 行rowname, 字段fieldname的值
func GetInt(tblname string, rowname interface{}, fieldname string) int32 {
	return _default_numbers.GetInt(tblname, rowname, fieldname)
}

func GetFloat(tblname string, rowname interface{}, fieldname string) float64 {
	return _default_numbers.GetFloat(tblname, rowname, fieldname)
}

func GetString(tblname string, rowname interface{}, fieldname string) string {
	return _default_numbers.GetString(tblname, rowname, fieldname)
}

// 读取表的所有KEY
func GetKeys(tblname string) []string {
	return _default_numbers.GetKeys(tblname)
}

// 返回表的总记录条数
func Count(tblname string) int32 {
	return _default_numbers.Count(tblname)
}

// 测试某个字段是否存在
func IsFieldExists(tblname string, fieldname string) bool {
	return _default_numbers.IsFieldExists(tblname, fieldname)
}

// 测试某条记录是否存在
func IsRecordExists(tblname string, rowname interface{}) bool {
	return _default_numbers.IsRecordExists(tblname, rowname)
}
