package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

const ExcelTimeFormat = time.RFC3339Nano

func FormatInt(v int64) string      { return strconv.FormatInt(v, 10) }
func FormatTime(v time.Time) string { return v.Format(ExcelTimeFormat) }

// ExcelRow decodes textual cells without converting money or IDs to floats.
type ExcelRow struct {
	Cells []string
	Err   error
}

func (r *ExcelRow) String(i int) string {
	if i >= len(r.Cells) {
		r.Err = fmt.Errorf("missing column %d", i+1)
		return ""
	}
	return r.Cells[i]
}
func (r *ExcelRow) Int(i int) int64 {
	v, err := strconv.ParseInt(r.String(i), 10, 64)
	if err != nil {
		r.Err = fmt.Errorf("column %d: %w", i+1, err)
	}
	return v
}
func (r *ExcelRow) Bool(i int) bool {
	v, err := strconv.ParseBool(r.String(i))
	if err != nil {
		r.Err = fmt.Errorf("column %d: %w", i+1, err)
	}
	return v
}
func (r *ExcelRow) Time(i int) time.Time {
	v, err := time.Parse(ExcelTimeFormat, r.String(i))
	if err != nil {
		r.Err = fmt.Errorf("column %d: %w", i+1, err)
	}
	return v
}

// SaveExcel writes beside the destination and replaces it only after success.
func SaveExcel(f *excelize.File, path string) (err error) {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".finance-*.xlsx")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = f.SaveAs(name); err != nil {
		return err
	}
	return os.Rename(name, path)
}
