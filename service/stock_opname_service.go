package service

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/go-sql-driver/mysql"

	"manisora/database"
	"manisora/model"
	"manisora/repository"
)

var ErrOpnameNotDraft = errors.New("opname sudah disetujui, tidak bisa diubah lagi")

func friendlyOpnameError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1452 {
		return errors.New("lokasi tidak ditemukan")
	}
	return err
}

func CreateStockOpname(o model.StockOpname) (model.StockOpname, error) {
	if o.LocationID <= 0 {
		return o, errors.New("location_id wajib diisi")
	}

	if o.DocDate == "" {
		o.DocDate = time.Now().Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", o.DocDate); err != nil {
		return o, errors.New("doc_date harus berformat YYYY-MM-DD")
	}
	o.Status = "DRAFT"

	tx, err := database.DB.Begin()
	if err != nil {
		return o, err
	}
	defer tx.Rollback()

	id, err := repository.CreateStockOpname(tx, o)
	if err != nil {
		return o, friendlyOpnameError(err)
	}

	docNo := fmt.Sprintf("OPN-%s-%04d", o.DocDate[:4], id)
	if err := repository.SetStockOpnameDocNo(tx, id, docNo); err != nil {
		return o, err
	}

	count, err := repository.SnapshotOpnameItems(tx, id, o.LocationID)
	if err != nil {
		return o, err
	}
	if count == 0 {
		return o, errors.New("lokasi ini belum punya data stok untuk di-opname")
	}

	if err := tx.Commit(); err != nil {
		return o, err
	}
	return GetStockOpnameByID(id)
}

func GetStockOpnameByID(id int64) (model.StockOpname, error) {
	o, err := repository.GetStockOpnameByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		return o, ErrNotFound
	}
	return o, err
}

func SaveOpnameCount(id int64, input model.CountInput) (model.StockOpname, error) {
	if len(input.Items) == 0 {
		return model.StockOpname{}, fmt.Errorf("%w: items tidak boleh kosong", ErrInvalidInput)
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return model.StockOpname{}, err
	}
	defer tx.Rollback()

	status, _, err := repository.LockStockOpname(tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.StockOpname{}, ErrNotFound
	}
	if err != nil {
		return model.StockOpname{}, err
	}
	if status != "DRAFT" {
		return model.StockOpname{}, ErrOpnameNotDraft
	}

	items, err := repository.GetStockOpnameItemsTx(tx, id)
	if err != nil {
		return model.StockOpname{}, err
	}
	inOpname := map[int64]bool{}
	for _, it := range items {
		inOpname[it.ItemID] = true
	}

	seen := map[int64]bool{}
	for _, c := range input.Items {
		if !inOpname[c.ItemID] {
			return model.StockOpname{}, fmt.Errorf("%w: barang id %d tidak ada dalam opname ini", ErrInvalidInput, c.ItemID)
		}
		if seen[c.ItemID] {
			return model.StockOpname{}, fmt.Errorf("%w: barang id %d muncul dua kali", ErrInvalidInput, c.ItemID)
		}
		if c.QtyPhysical < 0 {
			return model.StockOpname{}, fmt.Errorf("%w: qty_physical barang id %d tidak boleh negatif", ErrInvalidInput, c.ItemID)
		}
		seen[c.ItemID] = true
	}

	for _, c := range input.Items {
		if err := repository.SetOpnamePhysical(tx, id, c.ItemID, c.QtyPhysical); err != nil {
			return model.StockOpname{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.StockOpname{}, err
	}
	return GetStockOpnameByID(id)
}

func ApproveStockOpname(id int64) (model.StockOpname, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return model.StockOpname{}, err
	}
	defer tx.Rollback()

	status, locationID, err := repository.LockStockOpname(tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.StockOpname{}, ErrNotFound
	}
	if err != nil {
		return model.StockOpname{}, err
	}
	if status != "DRAFT" {
		return model.StockOpname{}, ErrOpnameNotDraft
	}

	items, err := repository.GetStockOpnameItemsTx(tx, id)
	if err != nil {
		return model.StockOpname{}, err
	}

	// Semua barang harus sudah dihitung sebelum disetujui
	for _, it := range items {
		if it.QtyPhysical == nil {
			return model.StockOpname{}, fmt.Errorf("%w: barang id %d belum dihitung", ErrInvalidInput, it.ItemID)
		}
	}

	for _, it := range items {
		diff := math.Round((*it.QtyPhysical-it.QtySystem)*1000) / 1000

		if err := repository.SetOpnameItemDiff(tx, it.ID, diff); err != nil {
			return model.StockOpname{}, err
		}
		if diff == 0 {
			continue
		}

		if diff < 0 {
			available, err := repository.GetStockForUpdate(tx, it.ItemID, locationID)
			if err != nil {
				return model.StockOpname{}, err
			}
			if available < -diff {
				return model.StockOpname{}, fmt.Errorf(
					"%w: barang id %d, saldo saat ini %.3f tidak cukup untuk penyesuaian %.3f, buat opname baru",
					ErrInsufficientStock, it.ItemID, available, diff)
			}
		}

		qtyIn, qtyOut := 0.0, 0.0
		if diff > 0 {
			qtyIn = diff
		} else {
			qtyOut = -diff
		}

		if err := repository.InsertLedger(tx, it.ItemID, locationID, qtyIn, qtyOut,
			"OPNAME_ADJUSTMENT", "STOCK_OPNAME", id); err != nil {
			return model.StockOpname{}, err
		}
		if err := repository.AddStockBalance(tx, it.ItemID, locationID, diff); err != nil {
			return model.StockOpname{}, err
		}
	}

	if err := repository.SetStockOpnameStatus(tx, id, "APPROVED"); err != nil {
		return model.StockOpname{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.StockOpname{}, err
	}
	return GetStockOpnameByID(id)
}
