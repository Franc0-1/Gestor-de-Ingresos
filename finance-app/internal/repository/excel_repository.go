package repository

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/utils"
	"github.com/xuri/excelize/v2"
)

const metadataSheet = "_Metadata"

var workbookLocks sync.Map

// ExcelRepository exposes the typed repositories, sharing one file lock.
// Multiple instances for the same absolute path also share the lock in this process.
type ExcelRepository struct {
	Accounts     AccountRepository
	Categories   CategoryRepository
	Transactions TransactionRepository
	Transfers    TransferRepository
	path         string
	mu           *sync.RWMutex
}

type sheetSpec struct {
	name    string
	headers []string
}

var sheets = []sheetSpec{
	{"Cuentas", []string{"ID", "Name", "Type", "Currency", "InitialBalance", "Active", "CreatedAt", "UpdatedAt"}},
	{"Categorias", []string{"ID", "Name", "Type", "CreatedAt", "UpdatedAt"}},
	{"Movimientos", []string{"ID", "Type", "Amount", "Description", "CategoryID", "AccountID", "Date", "CreatedAt", "UpdatedAt"}},
	{"Transferencias", []string{"ID", "FromAccountID", "ToAccountID", "Amount", "Description", "Date", "CreatedAt", "UpdatedAt"}},
}

// NewExcelRepository creates the parent directory and workbook if absent.
// No file handle remains open between operations.
func NewExcelRepository(path string) (*ExcelRepository, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	lock, _ := workbookLocks.LoadOrStore(abs, &sync.RWMutex{})
	r := &ExcelRepository{path: abs, mu: lock.(*sync.RWMutex)}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.initialize(); err != nil {
		return nil, err
	}
	r.Accounts = &excelTable[models.Account]{r, sheets[0], encodeAccount, decodeAccount}
	r.Categories = &excelTable[models.Category]{r, sheets[1], encodeCategory, decodeCategory}
	r.Transactions = &transactionTable{&excelTable[models.Transaction]{r, sheets[2], encodeTransaction, decodeTransaction}}
	r.Transfers = &transferTable{&excelTable[models.Transfer]{r, sheets[3], encodeTransfer, decodeTransfer}}
	return r, nil
}

func (r *ExcelRepository) initialize() (err error) {
	if err = os.MkdirAll(filepath.Dir(r.path), 0755); err != nil {
		return err
	}
	f, err := excelize.OpenFile(r.path)
	if err == nil {
		defer func() { err = errors.Join(err, f.Close()) }()
		// Validate the original sheets before upgrading an existing workbook.
		if err = validateWorkbook(f, sheets[:3]); err != nil {
			return err
		}
		changed, upgradeErr := ensureTransferSheet(f)
		if upgradeErr != nil {
			return upgradeErr
		}
		if err = validateWorkbook(f, sheets); err != nil {
			return err
		}
		if changed {
			return utils.SaveExcel(f, r.path)
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	f = excelize.NewFile()
	defer func() { err = errors.Join(err, f.Close()) }()
	for _, spec := range sheets {
		if _, err = f.NewSheet(spec.name); err != nil {
			return err
		}
		if err = writeRow(f, spec.name, 1, spec.headers); err != nil {
			return err
		}
	}
	if _, err = f.NewSheet(metadataSheet); err != nil {
		return err
	}
	if err = writeRow(f, metadataSheet, 1, []string{"Sheet", "LastID"}); err != nil {
		return err
	}
	for i, spec := range sheets {
		if err = writeRow(f, metadataSheet, i+2, []string{spec.name, "0"}); err != nil {
			return err
		}
	}
	if err = f.DeleteSheet("Sheet1"); err != nil {
		return err
	}
	f.SetActiveSheet(0)
	if err = f.SetSheetVisible(metadataSheet, false); err != nil {
		return err
	}
	return utils.SaveExcel(f, r.path)
}

func validateWorkbook(f *excelize.File, specs []sheetSpec) error {
	for _, spec := range specs {
		rows, err := f.GetRows(spec.name)
		if err != nil {
			return fmt.Errorf("%w: %s: %v", ErrInvalidWorkbook, spec.name, err)
		}
		if len(rows) == 0 || !slices.Equal(rows[0], spec.headers) {
			return fmt.Errorf("%w: headers in %s", ErrInvalidWorkbook, spec.name)
		}
	}
	for _, spec := range specs {
		if _, _, err := sequence(f, spec.name); err != nil {
			return err
		}
	}
	return nil
}

// Upgrade only the new sheet and sequence, preserving all existing cells and IDs.
func ensureTransferSheet(f *excelize.File) (bool, error) {
	spec := sheets[3]
	index, err := f.GetSheetIndex(spec.name)
	if err != nil {
		return false, err
	}
	changed := index == -1
	if changed {
		if _, err := f.NewSheet(spec.name); err != nil {
			return false, err
		}
		if err := writeRow(f, spec.name, 1, spec.headers); err != nil {
			return false, err
		}
	}
	metadata, err := f.GetRows(metadataSheet)
	if err != nil {
		return false, err
	}
	for _, row := range metadata {
		if len(row) > 0 && row[0] == spec.name {
			return changed, nil
		}
	}
	rows, err := f.GetRows(spec.name)
	if err != nil {
		return false, err
	}
	if len(rows) == 0 || !slices.Equal(rows[0], spec.headers) {
		return false, fmt.Errorf("%w: headers in %s", ErrInvalidWorkbook, spec.name)
	}
	var last int64
	for i, row := range rows[1:] {
		if len(row) == 0 {
			return false, fmt.Errorf("%w: empty transfer row %d", ErrInvalidWorkbook, i+2)
		}
		id, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil || id <= 0 {
			return false, fmt.Errorf("%w: transfer row %d ID", ErrInvalidWorkbook, i+2)
		}
		if id > last {
			last = id
		}
	}
	if err := writeRow(f, metadataSheet, len(metadata)+1, []string{spec.name, utils.FormatInt(last)}); err != nil {
		return false, err
	}
	return true, nil
}

func writeRow(f *excelize.File, sheet string, row int, values []string) error {
	cells := make([]interface{}, len(values))
	for i, v := range values {
		cells[i] = v
	}
	return f.SetSheetRow(sheet, fmt.Sprintf("A%d", row), &cells)
}

func sequence(f *excelize.File, sheet string) (int64, int, error) {
	rows, err := f.GetRows(metadataSheet)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: metadata: %v", ErrInvalidWorkbook, err)
	}
	for i, row := range rows {
		if len(row) == 2 && row[0] == sheet {
			id, err := strconv.ParseInt(row[1], 10, 64)
			if err != nil || id < 0 {
				return 0, 0, fmt.Errorf("%w: sequence in %s", ErrInvalidWorkbook, sheet)
			}
			return id, i + 1, nil
		}
	}
	return 0, 0, fmt.Errorf("%w: missing sequence for %s", ErrInvalidWorkbook, sheet)
}

type validModel interface{ Validate() error }

// excelTable shares CRUD mechanics while codecs keep each model explicit.
type excelTable[T validModel] struct {
	repo   *ExcelRepository
	spec   sheetSpec
	encode func(T) []string
	decode func([]string) (T, error)
}

func (t *excelTable[T]) withFile(write bool, fn func(*excelize.File) error) (err error) {
	if write {
		t.repo.mu.Lock()
		defer t.repo.mu.Unlock()
	} else {
		t.repo.mu.RLock()
		defer t.repo.mu.RUnlock()
	}
	f, err := excelize.OpenFile(t.repo.path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	if err = fn(f); err != nil {
		return err
	}
	if write {
		return utils.SaveExcel(f, t.repo.path)
	}
	return nil
}

func (t *excelTable[T]) rows(f *excelize.File) ([][]string, error) {
	rows, err := f.GetRows(t.spec.name)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 || !slices.Equal(rows[0], t.spec.headers) {
		return nil, fmt.Errorf("%w: %s headers", ErrInvalidWorkbook, t.spec.name)
	}
	for i, row := range rows[1:] {
		if len(row) != len(t.spec.headers) {
			return nil, fmt.Errorf("%w: %s row %d", ErrInvalidWorkbook, t.spec.name, i+2)
		}
		id, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("%w: %s row %d ID", ErrInvalidWorkbook, t.spec.name, i+2)
		}
	}
	return rows, nil
}

func (t *excelTable[T]) Create(value T) (result T, err error) {
	if err = value.Validate(); err != nil {
		return result, err
	}
	err = t.withFile(true, func(f *excelize.File) error {
		rows, err := t.rows(f)
		if err != nil {
			return err
		}
		last, metadataRow, err := sequence(f, t.spec.name)
		if err != nil {
			return err
		}
		// Include existing IDs to avoid collisions if the workbook was edited manually.
		for _, row := range rows[1:] {
			id, _ := strconv.ParseInt(row[0], 10, 64)
			if id > last {
				last = id
			}
		}
		if last == math.MaxInt64 {
			return ErrIDExhausted
		}
		cells := t.encode(value)
		cells[0] = utils.FormatInt(last + 1)
		now := utils.FormatTime(time.Now().UTC())
		zero := utils.FormatTime(time.Time{})
		for _, i := range []int{len(cells) - 2, len(cells) - 1} {
			if cells[i] == zero {
				cells[i] = now
			}
		}
		result, err = t.decode(cells)
		if err != nil {
			return err
		}
		if err = writeRow(f, t.spec.name, len(rows)+1, cells); err != nil {
			return err
		}
		return writeRow(f, metadataSheet, metadataRow, []string{t.spec.name, cells[0]})
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return result, nil
}

func (t *excelTable[T]) GetAll() (result []T, err error) {
	result = make([]T, 0)
	err = t.withFile(false, func(f *excelize.File) error {
		rows, err := t.rows(f)
		if err != nil {
			return err
		}
		for i, row := range rows[1:] {
			value, err := t.decode(row)
			if err != nil {
				return fmt.Errorf("%w: %s row %d: %v", ErrInvalidWorkbook, t.spec.name, i+2, err)
			}
			result = append(result, value)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func findRow(rows [][]string, id int64) (int, error) {
	for i := 1; i < len(rows); i++ {
		if rows[i][0] == utils.FormatInt(id) {
			return i, nil
		}
	}
	return 0, fmt.Errorf("ID %d: %w", id, ErrNotFound)
}

func (t *excelTable[T]) GetByID(id int64) (result T, err error) {
	err = t.withFile(false, func(f *excelize.File) error {
		rows, err := t.rows(f)
		if err != nil {
			return err
		}
		i, err := findRow(rows, id)
		if err != nil {
			return err
		}
		result, err = t.decode(rows[i])
		if err != nil {
			return fmt.Errorf("%w: %s: %v", ErrInvalidWorkbook, t.spec.name, err)
		}
		return nil
	})
	return result, err
}

func (t *excelTable[T]) Update(value T) error {
	if err := value.Validate(); err != nil {
		return err
	}
	return t.withFile(true, func(f *excelize.File) error {
		rows, err := t.rows(f)
		if err != nil {
			return err
		}
		cells := t.encode(value)
		id, _ := strconv.ParseInt(cells[0], 10, 64)
		i, err := findRow(rows, id)
		if err != nil {
			return err
		}
		cells[len(cells)-2] = rows[i][len(cells)-2]
		cells[len(cells)-1] = utils.FormatTime(time.Now().UTC())
		return writeRow(f, t.spec.name, i+1, cells)
	})
}

func (t *excelTable[T]) Delete(id int64) error {
	return t.withFile(true, func(f *excelize.File) error {
		rows, err := t.rows(f)
		if err != nil {
			return err
		}
		i, err := findRow(rows, id)
		if err != nil {
			return err
		}
		return f.RemoveRow(t.spec.name, i+1)
	})
}

func encodeAccount(a models.Account) []string {
	return []string{utils.FormatInt(a.ID), a.Name, a.Type, a.Currency, utils.FormatInt(a.InitialBalance), strconv.FormatBool(a.Active), utils.FormatTime(a.CreatedAt), utils.FormatTime(a.UpdatedAt)}
}
func decodeAccount(cells []string) (models.Account, error) {
	r := utils.ExcelRow{Cells: cells}
	a := models.Account{ID: r.Int(0), Name: r.String(1), Type: r.String(2), Currency: r.String(3), InitialBalance: r.Int(4), Active: r.Bool(5), CreatedAt: r.Time(6), UpdatedAt: r.Time(7)}
	return a, r.Err
}
func encodeCategory(c models.Category) []string {
	return []string{utils.FormatInt(c.ID), c.Name, c.Type, utils.FormatTime(c.CreatedAt), utils.FormatTime(c.UpdatedAt)}
}
func decodeCategory(cells []string) (models.Category, error) {
	r := utils.ExcelRow{Cells: cells}
	c := models.Category{ID: r.Int(0), Name: r.String(1), Type: r.String(2), CreatedAt: r.Time(3), UpdatedAt: r.Time(4)}
	return c, r.Err
}
func encodeTransaction(t models.Transaction) []string {
	return []string{utils.FormatInt(t.ID), t.Type, utils.FormatInt(t.Amount), t.Description, utils.FormatInt(t.CategoryID), utils.FormatInt(t.AccountID), utils.FormatTime(t.Date), utils.FormatTime(t.CreatedAt), utils.FormatTime(t.UpdatedAt)}
}
func decodeTransaction(cells []string) (models.Transaction, error) {
	r := utils.ExcelRow{Cells: cells}
	t := models.Transaction{ID: r.Int(0), Type: r.String(1), Amount: r.Int(2), Description: r.String(3), CategoryID: r.Int(4), AccountID: r.Int(5), Date: r.Time(6), CreatedAt: r.Time(7), UpdatedAt: r.Time(8)}
	return t, r.Err
}

var (
	_ AccountRepository     = (*excelTable[models.Account])(nil)
	_ CategoryRepository    = (*excelTable[models.Category])(nil)
	_ TransactionRepository = (*transactionTable)(nil)
	_ TransferRepository    = (*transferTable)(nil)
)

func encodeTransfer(t models.Transfer) []string {
	return []string{utils.FormatInt(t.ID), utils.FormatInt(t.FromAccountID), utils.FormatInt(t.ToAccountID), utils.FormatInt(t.Amount), t.Description, utils.FormatTime(t.Date), utils.FormatTime(t.CreatedAt), utils.FormatTime(t.UpdatedAt)}
}

func decodeTransfer(cells []string) (models.Transfer, error) {
	r := utils.ExcelRow{Cells: cells}
	t := models.Transfer{ID: r.Int(0), FromAccountID: r.Int(1), ToAccountID: r.Int(2), Amount: r.Int(3), Description: r.String(4), Date: r.Time(5), CreatedAt: r.Time(6), UpdatedAt: r.Time(7)}
	return t, r.Err
}
