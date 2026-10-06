package repository

import (
	"errors"
	"math"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"finance-app/internal/models"
	"github.com/xuri/excelize/v2"
)

func newTestRepository(t *testing.T) (*ExcelRepository, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data", "finanzas.xlsx")
	r, err := NewExcelRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	return r, path
}

func TestWorkbookCreation(t *testing.T) {
	_, path := newTestRepository(t)
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})
	for _, spec := range sheets {
		rows, err := f.GetRows(spec.name)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 || !slices.Equal(rows[0], spec.headers) {
			t.Fatalf("incorrect headers in %s: %v", spec.name, rows)
		}
	}
	if slices.Contains(f.GetSheetList(), "Sheet1") {
		t.Fatal("Sheet1 was not removed")
	}
	visible, err := f.GetSheetVisible(metadataSheet)
	if err != nil || visible {
		t.Fatalf("metadata visibility: %v, %v", visible, err)
	}
	if _, err := NewExcelRepository(path); err != nil {
		t.Fatal(err)
	}
}

type testCRUD[T any] interface {
	Create(T) (T, error)
	GetAll() ([]T, error)
	GetByID(int64) (T, error)
	Update(T) error
	Delete(int64) error
}

func exerciseCRUD[T validModel](t *testing.T, repo testCRUD[T], reopen func() testCRUD[T], sample T, encode func(T) []string, decode func([]string) (T, error), editColumn int) {
	t.Helper()
	all, err := repo.GetAll()
	if err != nil || len(all) != 0 {
		t.Fatalf("initial records: %v %v", all, err)
	}
	first, err := repo.Create(sample)
	if err != nil {
		t.Fatal(err)
	}
	if encode(first)[0] != "1" {
		t.Fatal("first ID must be 1")
	}
	second, err := repo.Create(sample)
	if err != nil {
		t.Fatal(err)
	}
	if encode(second)[0] != "2" {
		t.Fatal("second ID must be 2")
	}
	repo = reopen()
	got, err := repo.GetByID(1)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(encode(got), encode(first)) {
		t.Fatalf("round trip mismatch: %v / %v", encode(got), encode(first))
	}
	all, err = repo.GetAll()
	if err != nil || len(all) != 2 {
		t.Fatalf("GetAll: %v %v", all, err)
	}
	cells := encode(first)
	cells[editColumn] = "updated"
	created := cells[len(cells)-2]
	cells[len(cells)-2] = time.Time{}.Format(time.RFC3339Nano)
	updated, err := decode(cells)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(updated); err != nil {
		t.Fatal(err)
	}
	got, err = reopen().GetByID(1)
	if err != nil {
		t.Fatal(err)
	}
	gotCells := encode(got)
	if gotCells[editColumn] != "updated" || gotCells[len(gotCells)-2] != created {
		t.Fatal("update did not preserve fields and CreatedAt")
	}
	stamp, err := time.Parse(time.RFC3339Nano, gotCells[len(gotCells)-1])
	if err != nil || stamp.IsZero() {
		t.Fatal("missing UpdatedAt")
	}
	if err := repo.Delete(2); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(1); err != nil {
		t.Fatal(err)
	}
	repo = reopen()
	for _, id := range []int64{1, 999, 0, -1} {
		if _, err := repo.GetByID(id); !errors.Is(err, ErrNotFound) {
			t.Fatalf("GetByID(%d): %v", id, err)
		}
		if err := repo.Delete(id); !errors.Is(err, ErrNotFound) {
			t.Fatalf("Delete(%d): %v", id, err)
		}
	}
	if err := repo.Update(updated); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update missing: %v", err)
	}
	third, err := repo.Create(sample)
	if err != nil {
		t.Fatal(err)
	}
	if encode(third)[0] != "3" {
		t.Fatal("deleted IDs reused after reopening")
	}
	all, err = repo.GetAll()
	if err != nil || len(all) != 1 {
		t.Fatalf("records after deletion: %v %v", all, err)
	}
}

func TestCRUD(t *testing.T) {
	stamp := time.Date(2024, 2, 3, 4, 5, 6, 123456789, time.FixedZone("ART", -3*60*60))
	t.Run("Account", func(t *testing.T) {
		r, path := newTestRepository(t)
		reopen := func() testCRUD[models.Account] {
			r, err := NewExcelRepository(path)
			if err != nil {
				t.Fatal(err)
			}
			return r.Accounts
		}
		a := models.Account{Name: "Caja", Type: "cash", Currency: "ARS", InitialBalance: math.MinInt64, Active: true, CreatedAt: stamp, UpdatedAt: stamp}
		exerciseCRUD(t, r.Accounts, reopen, a, encodeAccount, decodeAccount, 1)
		a.Active = false
		a.InitialBalance = math.MaxInt64
		saved, err := r.Accounts.Create(a)
		if err != nil {
			t.Fatal(err)
		}
		got, err := reopen().GetByID(saved.ID)
		if err != nil || got.Active || got.InitialBalance != math.MaxInt64 || !got.CreatedAt.Equal(stamp) {
			t.Fatalf("false/int64/time persistence: %+v %v", got, err)
		}
	})
	t.Run("Category", func(t *testing.T) {
		r, path := newTestRepository(t)
		reopen := func() testCRUD[models.Category] {
			r, err := NewExcelRepository(path)
			if err != nil {
				t.Fatal(err)
			}
			return r.Categories
		}
		exerciseCRUD(t, r.Categories, reopen, models.Category{Name: "Sueldo", Type: "income", CreatedAt: stamp, UpdatedAt: stamp}, encodeCategory, decodeCategory, 1)
	})
	t.Run("Transaction", func(t *testing.T) {
		r, path := newTestRepository(t)
		reopen := func() testCRUD[models.Transaction] {
			r, err := NewExcelRepository(path)
			if err != nil {
				t.Fatal(err)
			}
			return r.Transactions
		}
		v := models.Transaction{Type: "expense", Amount: math.MaxInt64, Description: "=literal", AccountID: math.MaxInt64, CategoryID: math.MaxInt64, Date: stamp, CreatedAt: stamp, UpdatedAt: stamp}
		exerciseCRUD(t, r.Transactions, reopen, v, encodeTransaction, decodeTransaction, 3)
		got, err := reopen().GetByID(3)
		if err != nil || !got.Date.Equal(stamp) || got.Date.Nanosecond() != stamp.Nanosecond() {
			t.Fatalf("Date persistence: %+v %v", got, err)
		}
	})
}

func TestRepositoryValidation(t *testing.T) {
	r, _ := newTestRepository(t)
	_, err := r.Accounts.Create(models.Account{})
	var validation *models.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("expected typed validation error: %v", err)
	}
	_, err = r.Categories.Create(models.Category{})
	if !errors.As(err, &validation) {
		t.Fatalf("expected category validation: %v", err)
	}
	_, err = r.Transactions.Create(models.Transaction{})
	if !errors.As(err, &validation) {
		t.Fatalf("expected transaction validation: %v", err)
	}
	a, err := r.Accounts.Create(models.Account{Name: "Cash", Type: "cash", Currency: "ARS"})
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != 1 || a.CreatedAt.IsZero() || a.UpdatedAt.IsZero() {
		t.Fatal("invalid creates consumed IDs or timestamps absent")
	}
	a.Currency = ""
	if err := r.Accounts.Update(a); !errors.As(err, &validation) {
		t.Fatalf("expected update validation: %v", err)
	}
}

func TestConcurrentRepositories(t *testing.T) {
	r, path := newTestRepository(t)
	other, err := NewExcelRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			repo := r
			if i%2 == 0 {
				repo = other
			}
			if _, err := repo.Categories.Create(models.Category{Name: strconv.Itoa(i), Type: "expense"}); err != nil {
				t.Error(err)
			}
			if _, err := repo.Categories.GetAll(); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	all, err := r.Categories.GetAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 12 {
		t.Fatalf("lost writes: %d", len(all))
	}
	for i, c := range all {
		if c.ID != int64(i+1) {
			t.Fatalf("unexpected ID: %d", c.ID)
		}
	}
}

func TestInvalidWorkbookAndIDExhaustion(t *testing.T) {
	r, path := newTestRepository(t)
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeRow(f, metadataSheet, 2, []string{"Cuentas", strconv.FormatInt(math.MaxInt64, 10)}); err != nil {
		t.Fatal(err)
	}
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Accounts.Create(models.Account{Name: "Cash", Type: "cash", Currency: "ARS"}); !errors.Is(err, ErrIDExhausted) {
		t.Fatalf("expected exhaustion: %v", err)
	}
	f, err = excelize.OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellStr("Cuentas", "A1", "Broken"); err != nil {
		t.Fatal(err)
	}
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := NewExcelRepository(path); !errors.Is(err, ErrInvalidWorkbook) {
		t.Fatalf("expected invalid workbook: %v", err)
	}
}
