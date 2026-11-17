package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"

	"manisora/database"
	"manisora/model"
	"manisora/repository"
)

var ErrReturnExceedsReceipt = errors.New("jumlah retur melebihi jumlah yang diterima")

func friendlyReturnError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1452 {
		return errors.New("supplier, lokasi, penerimaan, atau barang tidak ditemukan")
	}
	return err
}

func CreateSupplierReturn(r model.SupplierReturn) (model.SupplierReturn, error) {
	if r.SupplierID <= 0 || r.LocationID <= 0 {
		return r, errors.New("supplier_id dan location_id wajib diisi")
	}
	if r.ReceiptID < 0 {
		return r, errors.New("receipt_id tidak valid")
	}
	if len(r.Items) == 0 {
		return r, errors.New("minimal satu barang harus diisi")
	}

	seen := map[int64]bool{}
	for _, it := range r.Items {
		if it.ItemID <= 0 || it.Qty <= 0 {
			return r, errors.New("setiap barang harus punya item_id dan qty lebih dari 0")
		}
		if seen[it.ItemID] {
			return r, errors.New("barang yang sama tidak boleh muncul dua kali dalam satu retur")
		}
		seen[it.ItemID] = true
	}

	if r.DocDate == "" {
		r.DocDate = time.Now().Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", r.DocDate); err != nil {
		return r, errors.New("doc_date harus berformat YYYY-MM-DD")
	}
	r.Status = "DRAFT"

	tx, err := database.DB.Begin()
	if err != nil {
		return r, err
	}
	defer tx.Rollback()

	id, err := repository.CreateSupplierReturn(tx, r)
	if err != nil {
		return r, friendlyReturnError(err)
	}
	r.ID = id
	r.DocNo = fmt.Sprintf("RTR-%s-%04d", r.DocDate[:4], id)

	if err := repository.SetSupplierReturnDocNo(tx, id, r.DocNo); err != nil {
		return r, err
	}

	for i := range r.Items {
		itemID, err := repository.CreateSupplierReturnItem(tx, id, r.Items[i])
		if err != nil {
			return r, friendlyReturnError(err)
		}
		r.Items[i].ID = itemID
		r.Items[i].ReturnID = id
	}

	if err := tx.Commit(); err != nil {
		return r, err
	}
	return r, nil
}

func GetSupplierReturnByID(id int64) (model.SupplierReturn, error) {
	r, err := repository.GetSupplierReturnByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		return r, ErrNotFound
	}
	return r, err
}

func PostSupplierReturn(id int64) (model.SupplierReturn, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return model.SupplierReturn{}, err
	}
	defer tx.Rollback()

	status, supplierID, locationID, receiptID, err := repository.LockSupplierReturn(tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.SupplierReturn{}, ErrNotFound
	}
	if err != nil {
		return model.SupplierReturn{}, err
	}
	if status != "DRAFT" {
		return model.SupplierReturn{}, ErrNotDraft
	}

	items, err := repository.GetSupplierReturnItemsTx(tx, id)
	if err != nil {
		return model.SupplierReturn{}, err
	}
	if len(items) == 0 {
		return model.SupplierReturn{}, errors.New("dokumen tidak punya barang")
	}

	// Aturan 1: kalau ada penerimaan acuan, retur tidak boleh melebihi yang diterima
	if receiptID > 0 {
		rStatus, rSupplierID, err := repository.LockReceiptForReturn(tx, receiptID)
		if err != nil {
			return model.SupplierReturn{}, err
		}
		if rStatus != "POSTED" {
			return model.SupplierReturn{}, fmt.Errorf("%w: penerimaan acuan belum di-post", ErrInvalidInput)
		}
		if rSupplierID != supplierID {
			return model.SupplierReturn{}, fmt.Errorf("%w: supplier retur berbeda dengan supplier pada penerimaan acuan", ErrInvalidInput)
		}

		for _, it := range items {
			received, err := repository.GetReceivedQty(tx, receiptID, it.ItemID)
			if err != nil {
				return model.SupplierReturn{}, err
			}
			returned, err := repository.GetReturnedQty(tx, receiptID, it.ItemID)
			if err != nil {
				return model.SupplierReturn{}, err
			}
			if it.Qty > received-returned {
				return model.SupplierReturn{}, fmt.Errorf(
					"%w: barang id %d, diterima %.3f, sudah diretur %.3f, diminta %.3f",
					ErrReturnExceedsReceipt, it.ItemID, received, returned, it.Qty)
			}
		}
	}

	// Aturan 2: stok di lokasi harus cukup, lalu stok keluar
	for _, it := range items {
		available, err := repository.GetStockForUpdate(tx, it.ItemID, locationID)
		if err != nil {
			return model.SupplierReturn{}, err
		}
		if available < it.Qty {
			return model.SupplierReturn{}, fmt.Errorf("%w: barang id %d, tersedia %.3f, diminta %.3f",
				ErrInsufficientStock, it.ItemID, available, it.Qty)
		}

		if err := repository.InsertLedger(tx, it.ItemID, locationID, 0, it.Qty,
			"SUPPLIER_RETURN", "SUPPLIER_RETURN", id); err != nil {
			return model.SupplierReturn{}, err
		}
		if err := repository.AddStockBalance(tx, it.ItemID, locationID, -it.Qty); err != nil {
			return model.SupplierReturn{}, err
		}
	}

	if err := repository.SetSupplierReturnStatus(tx, id, "POSTED"); err != nil {
		return model.SupplierReturn{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.SupplierReturn{}, err
	}
	return GetSupplierReturnByID(id)
}
