package numbers

import (
	"fmt"
	"testing"
)

func TestNumbers(t *testing.T) {
	/*
		excelFileName := "test.xlsx"
		xlFile, err := xlsx.OpenFile(excelFileName)
		if err != nil {
			t.Fatal(err)
		}
		_default_numbers.parse(path.Base(excelFileName), xlFile.Sheets)
		spew.Dump(_default_numbers)
	*/
	ns := Numbers("TaskConfig")
	keys := ns.GetKeys("KEY_Achievement")
	for k, v := range keys {
		fmt.Println(k, v)
	}

	config := Numbers("GachaCfg")
	fmt.Println(config.IsRecordExists("KEY_Data", 1))

}

func TestWatch(t *testing.T) {
	/*
		ns := <-CH()
		for k, v := range ns.tables {
			fmt.Println(k, *v)
		}
	*/
}

/*
func Benchmark_init(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n := numbers{}
		n.init(DEFAULT_NUMBERS_PATH)
	}
}
*/
