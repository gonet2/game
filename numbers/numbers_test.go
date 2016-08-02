package numbers

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/tealeg/xlsx"
	"testing"
)

func TestNumbers(t *testing.T) {
	excelFileName := "test.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		t.Fatal(err)
	}
	_default_numbers.parse(xlFile.Sheets)
	spew.Dump(_default_numbers)
}
