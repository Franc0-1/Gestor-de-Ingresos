package repository

import (
	"bytes"
	"errors"
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	"finance-app/internal/models"
	"github.com/xuri/excelize/v2"
)

func TestTransferCRUD(t *testing.T) {
	r, path := newTestRepository(t)
	reopen := func() testCRUD[models.Transfer] {
		r, err := NewExcelRepository(path)
		if err != nil {
			t.Fatal(err)
		}
		return r.Transfers
	}
	stamp := time.Date(2026, 10, 6, 0, 0, 0, 123456789, time.FixedZone("ART", -3*60*60))
	sample := models.Transfer{FromAccountID: math.MaxInt64, ToAccountID: math.MaxInt64 - 1, Amount: math.MaxInt64, Description: "Mover dinero", Date: stamp, CreatedAt: stamp, UpdatedAt: stamp}
	exerciseCRUD(t, r.Transfers, reopen, sample, encodeTransfer, decodeTransfer, 4)
	got, err := reopen().GetByID(3)
	if err != nil || got.Amount != math.MaxInt64 || got.FromAccountID != math.MaxInt64 || got.ToAccountID != math.MaxInt64-1 || !got.Date.Equal(stamp) || got.Date.Nanosecond() != stamp.Nanosecond() || !got.CreatedAt.Equal(stamp) {
		t.Fatalf("transfer precision: %+v %v", got, err)
	}
	transactions, err := r.Transactions.GetAll()
	if err != nil || len(transactions) != 0 {
		t.Fatalf("transfer created transactions: %v %v", transactions, err)
	}
}

func TestTransferWorkbookUpgrade(t *testing.T) {
	r, path := newTestRepository(t)
	a, err := r.Accounts.Create(models.Account{Name: "Cash", Type: "cash", Currency: "ARS", InitialBalance: math.MinInt64, Active: true})
	if err != nil {
		t.Fatal(err)
	}
	deleted, err := r.Accounts.Create(models.Account{Name: "Deleted", Type: "cash", Currency: "ARS"})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Accounts.Delete(deleted.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Categories.Create(models.Category{Name: "Salary", Type: "income"}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Transactions.Create(models.Transaction{Type: "income", Amount: math.MaxInt64, AccountID: a.ID, CategoryID: 1, Date: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, metaRow, err := sequence(f, "Transferencias")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.DeleteSheet("Transferencias"); err != nil {
		t.Fatal(err)
	}
	if err := f.RemoveRow(metadataSheet, metaRow); err != nil {
		t.Fatal(err)
	}
	before := make(map[string][][]string)
	for _, sheet := range []string{"Cuentas", "Categorias", "Movimientos", metadataSheet} {
		rows, err := f.GetRows(sheet)
		if err != nil {
			t.Fatal(err)
		}
		before[sheet] = rows
	}
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	r, err = NewExcelRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	f, err = excelize.OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for sheet, expected := range before {
		actual, err := f.GetRows(sheet)
		if err != nil {
			t.Fatal(err)
		}
		if sheet == metadataSheet {
			actual = actual[:len(expected)]
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("upgrade changed %s", sheet)
		}
	}
	if err := validateWorkbook(f, sheets); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewExcelRepository(path); err != nil {
		t.Fatal(err)
	}
	again, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(content, again) {
		t.Fatalf("second initialization rewrote workbook: %v", err)
	}
	next, err := r.Accounts.Create(models.Account{Name: "Next", Type: "cash", Currency: "ARS"})
	if err != nil || next.ID != 3 {
		t.Fatalf("account sequence reset: %+v %v", next, err)
	}
	transfer, err := r.Transfers.Create(models.Transfer{FromAccountID: 1, ToAccountID: 3, Amount: 500, Date: time.Now().UTC()})
	if err != nil || transfer.ID != 1 {
		t.Fatalf("new transfer sequence: %+v %v", transfer, err)
	}
}

func TestTransferSequenceRecovery(t *testing.T) {
	r, path := newTestRepository(t)
	tx, err := r.Transfers.Create(models.Transfer{FromAccountID: 1, ToAccountID: 2, Amount: 100, Date: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, row, err := sequence(f, "Transferencias")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.RemoveRow(metadataSheet, row); err != nil {
		t.Fatal(err)
	}
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	r, err = NewExcelRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Transfers.Delete(tx.ID); err != nil {
		t.Fatal(err)
	}
	r, err = NewExcelRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	next, err := r.Transfers.Create(tx)
	if err != nil || next.ID != tx.ID+1 {
		t.Fatalf("recovered sequence reused ID: %+v %v", next, err)
	}
}

func TestTransferUpgradeRejectsInvalidSheet(t *testing.T) {
	_, path := newTestRepository(t)
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, row, err := sequence(f, "Transferencias")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.RemoveRow(metadataSheet, row); err != nil {
		t.Fatal(err)
	}
	if err := f.SetCellStr("Transferencias", "A1", "Broken"); err != nil {
		t.Fatal(err)
	}
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewExcelRepository(path); !errors.Is(err, ErrInvalidWorkbook) {
		t.Fatalf("expected invalid workbook: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("invalid workbook was overwritten: %v", err)
	}
}
